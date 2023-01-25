# Golang Load Balancer
Project of a Load Balancer in Golang

A load balancer is a device responsible for distributing network traffic among a certain number of servers. Its a fundamental piece to increase reliability and robustness of modern applications

The concept behind this project was based upon the , it consists of a basic load balancer written in Golang.

The input is given through a configuration file reffered as `config.yaml`,  example of a configuration file down below:

```yaml
## Load Balancer Port
lb_port: 3333
## Retry limit of a backend server 
retry_limit: 3
## Backend servers host
backends:
  - "http://localhost:100"
  - "http://localhost:101"
```

This load balancer supports two strategies:
- Round Robin: this algorithm equally distributes incoming traffic among the available servers
- Least Connections: traffic is distributed taking into account the number of active connections in each server

The strategy con be specified in the configuration file:

```yaml
lb_port: 3333
## LB strategy (uses round-robin as default)
strategy: least-connection
retry_limit: 3
backends:
  - "http://localhost:100"
  - "http://localhost:101"
```
## References

- [simplelb](https://github.com/kasvith/simplelb) repository project
- [Goconcurrency](https://gist.github.com/rushilgupta/228dfdf379121cb9426d5e90d34c5b96) gist
