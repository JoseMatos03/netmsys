# Network Monitoring System (netmsys)

Este projeto é uma aplicação modular para monitorização de agentes numa rede, com suporte a comunicação confiável e alertas em tempo real. A aplicação utiliza dois protocolos personalizados para comunicação entre agentes e servidor: **NetTask** (UDP) e **AlertFlow** (TCP).

## Funcionalidades

- **Monitorização de métricas**: Recolha de dados da rede a partir de agentes.
- **Sistema de alertas**: Notificações em tempo real baseadas em condições predefinidas.
- **Modularidade**: Arquitetura com separação clara entre cliente, servidor e protocolos, permitindo reusabilidade em outros projetos.
- **Suporte a falhas de rede**: Retransmissão de pacotes para garantir comunicação confiável.

## Arquitetura

A aplicação é composta por:

1. **Servidor**: Responsável por centralizar as métricas e alertas enviados pelos agentes.
2. **Agentes**: Dispositivos que monitorizam a rede e recursos locais e enviam os dados ao servidor.
3. **Protocolos personalizados**:
   - **NetTask**: Gerencia a comunicação UDP para envio de tarefas e receção de dados.
   - **AlertFlow**: Utiliza TCP para alertas mais confiáveis.

## Tecnologias Utilizadas

- **Linguagem**: Go
- **Bibliotecas**:
  - [`net`](https://pkg.go.dev/net): Comunicação UDP e TCP.
  - [`gopsutil`](https://github.com/shirou/gopsutil): Monitorização de métricas do sistema.
