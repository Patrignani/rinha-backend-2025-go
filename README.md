
<h1 align="center">
  <img src="./rinha_go.png" width="280" alt="Rinha Backend Logo"/>
  <br>
  Rinha de Backend 2025
</h1>

---

## ⚙️ Tecnologias Utilizadas

- **Linguagem:** Golang  
- **Servidor de alta performance:** [GNET](https://gnet.host/) (event-driven, multicore)  
- **Balanceador de carga:** [HAProxy](https://www.haproxy.org/)  
- **Banco de dados:** PostgreSQL  
- **Serialização rápida:** [EasyJSON](https://github.com/mailru/easyjson) (para desempenho na serialização/deserialização JSON)  
- **Servidor HTTP de alta performance:** [fasthttp](https://github.com/valyala/fasthttp) (para integração eficiente com APIs)  

---

## 📦 Endpoints da API

| Método | Rota                | Descrição                        |
|--------|---------------------|---------------------------------|
| POST   | `/payments`         | Cria um novo pagamento           |
| GET    | `/payments-summary` | Consulta histórico de pagamentos |

---

## ⚡ Comentários sobre as tecnologias usadas

- Utilizamos **EasyJSON** para serialização/deserialização JSON eficiente, garantindo alta performance na manipulação de payloads da API.  
- O servidor HTTP é baseado em **fasthttp**, que oferece menor latência e maior throughput comparado ao net/http padrão.  
- O GNET permite um modelo de servidor TCP **event-driven** com suporte multicore, ideal para alta concorrência em microserviços.  
- O balanceamento de carga é feito com **HAProxy**, garantindo distribuição eficiente das requisições entre múltiplas instâncias do serviço.

---

