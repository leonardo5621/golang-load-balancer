# Golang Load Balancer
Project of a Load Balancer in Golang

A load balancer is a device responsible for distributing network traffic among a certain number of servers. Its a fundamental piece to increase reliability and robustness of modern applications

The concept behind this project was based upon the [simplelb](https://github.com/kasvith/simplelb) repository project, it consists of a basic load balancer written in Golang.

The input is given through a configuration file reffered as `config.yaml`,  example of a configuration file down below:

```
## Load Balancer Port
lb_port: 3333
## In case of failed request, maximum amount of attempts in other hosts
max_attempt_limit: 3
## Retry limit of a backend server 
retry_limit: 3
## Backend servers host
backends:
  - "http://localhost:100"
  - "http://localhost:101"
```
