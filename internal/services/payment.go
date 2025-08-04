package services

import (
	"context"
	"fmt"
	"time"
	"unsafe"

	"github.com/Patrignani/patrignani-rinha-backend-go/internal/repositories"
	"github.com/Patrignani/patrignani-rinha-backend-go/internal/workers"
	"github.com/Patrignani/patrignani-rinha-backend-go/pkg/config"
	"github.com/Patrignani/patrignani-rinha-backend-go/pkg/models"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
)

// var randPool = sync.Pool{
// 	New: func() interface{} {
// 		return rand.New(rand.NewSource(time.Now().UnixNano()))
// 	},
// }

type PaymentService struct {
	queue        *workers.QueueWorker
	repo         *repositories.PaymentRepository
	defaultFast  *fasthttp.HostClient
	fallbackFast *fasthttp.HostClient
}

func NewPaymentService(repo *repositories.PaymentRepository, queue *workers.QueueWorker) *PaymentService {
	var fastClient1 = &fasthttp.HostClient{
		Addr:     config.Env.DefaultUrl,
		MaxConns: 2048,
	}

	var fastClient2 = &fasthttp.HostClient{
		Addr:     config.Env.FallbackUrl,
		MaxConns: 2048,
	}

	return &PaymentService{repo: repo, queue: queue, defaultFast: fastClient1, fallbackFast: fastClient2}
}

func (p *PaymentService) RunQueue(ctx context.Context, msg []byte) error {
	var parser fastjson.Parser
	v, err := parser.ParseBytes(msg)
	if err != nil {
		println(fmt.Sprintf("ERRO DE CONVERSÃO %v", err))
		return err
	}

	correlationId := b2s(v.GetStringBytes("correlationId"))
	amount := v.GetFloat64("amount")
	createdAt := time.Now().UTC()

	// if err := p.CallbackExc(ctx, correlationId, amount, createdAt, 0); err != nil {
	// 	if err := p.ExecuteFallback(ctx, correlationId, amount, createdAt); err != nil {
	// 		p.queue.Send(msg)
	// 		return err
	// 	}
	// }

	// if err := p.ExecuteDefault(ctx, correlationId, amount, createdAt); err != nil {
	// 	r := randPool.Get().(*rand.Rand)
	// 	defer randPool.Put(r)
	// 	if r.Intn(2) == 0 {
	// 		if err := p.ExecuteFallback(ctx, correlationId, amount, createdAt); err != nil {
	// 			p.queue.Send(msg)
	// 		}
	// 	}

	// 	p.queue.Send(msg)
	// }

	if err := p.ExecuteDefault(ctx, correlationId, amount, createdAt); err != nil {
		if err := p.ExecuteFallback(ctx, correlationId, amount, createdAt); err != nil {
			p.queue.Send(msg)
		}
	}

	return nil
}

// func (p *PaymentService) CallbackExc(ctx context.Context, correlationId string, amount float64, createdAt time.Time, attempts int) error {
// 	if err := p.ExecuteDefault(ctx, correlationId, amount, createdAt); err != nil {
// 		if attempts < config.Env.AttempsRetry {
// 			time.Sleep(config.Env.TimeAttemps)
// 			attempts++
// 			return p.CallbackExc(ctx, correlationId, amount, createdAt, attempts)
// 		}

// 		return err
// 	}

// 	return nil
// }

func (p *PaymentService) ExecuteDefault(ctx context.Context, correlationId string, amount float64, createdAt time.Time) error {
	return p.execute(ctx, correlationId, amount, createdAt, false)
}

func (p *PaymentService) ExecuteFallback(ctx context.Context, correlationId string, amount float64, createdAt time.Time) error {
	return p.execute(ctx, correlationId, amount, createdAt, true)
}

func (p *PaymentService) execute(ctx context.Context, correlationId string, amount float64, createdAt time.Time, fallback bool) error {

	statusCode, err := p.postPayment(fallback, correlationId, amount, createdAt)
	if err != nil {
		println(fmt.Sprintf("Erro post: %v", err))
		return err
	}

	if statusCode >= 200 && statusCode < 300 {
		go p.repo.Insert(ctx, models.PaymentDb{
			CorrelationId: correlationId,
			Amount:        amount,
			Fallback:      fallback,
			CreatedAt:     createdAt,
		})

		return nil
	} else if statusCode == 422 {
		return nil
	}

	return fmt.Errorf("HTTP status fora da faixa 2xx: %d", statusCode)
}

func (p *PaymentService) postPayment(fallback bool, correlationId string, amount float64, createdAt time.Time) (int, error) {

	pay := models.PaymentRequest{
		CorrelationId: correlationId,
		Amount:        amount,
		RequestedAt:   createdAt,
	}

	body, err := pay.MarshalJSON()

	if err != nil {
		println(fmt.Sprintf("HTTP status fora da faixa 2xx: %v", err))
	}

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI("/payments")
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.SetBodyRaw(body)

	if fallback {
		req.Header.Set("Host", config.Env.FallbackUrl)
		err := p.fallbackFast.Do(req, resp)
		return resp.StatusCode(), err
	}

	req.Header.Set("Host", config.Env.DefaultUrl)
	err = p.defaultFast.Do(req, resp)
	return resp.StatusCode(), err
}

func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func (p *PaymentService) GetPaymentSummary(ctx context.Context, from, to *time.Time) (*models.SummaryResponse, error) {
	return p.repo.GetPaymentSummary(ctx, from, to)
}
