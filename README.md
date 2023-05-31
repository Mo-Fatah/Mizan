# Mizan
Experimental Load Balancer to balance traffic between multiple services and multiple replicas of a service.
> Mizan is the Arabic word for Balance, pronounced as /mɪˈzæn/ or /mɪˈzɑːn/.

## Features
- **Multiple Balancing Algorithms**
    - Round Robin
    - Weighted Round Robin
    - *Upcoming: Random, Least Response Time, Least Connections*
- **Continuous Health Check**
    - Actively checking the health for each replica of each service. Directing traffic only to healthy replicas.

- **Hot Configuration Reloading**
    - Reloading configuration without restarting the load balancer with zero downtime.

- **Layer 7 Load Balancing**
    - Load balancing based on HTTP request path.

- **Graceful Shutdown**
    - Gracefully shutting down the load balancer without dropping any connections.

## Usage
### Configuration
Mizan uses YAML for configuration.  
a sample configuration file for a Round Robin balancer as follows:
```yaml
strategy: "rr"
max_connections: 1024 
ports:
  - 8080
  - 8081
  - 8082
services: 
  - matcher: "/api/v1"
    name: "test service"
    replicas:
      - url: "http://localhost:9090"
      - url: "http://localhost:9091"
      - url: "http://localhost:9092"
```
The configuration file is divided into 4 sections:
- **strategy**: the load balancing strategy to use. currently only round robin is supported.
- **max_connections**: the maximum number of connections to be handled by the load balancer.
- **ports**: the ports to listen on.
- **services**: the services to be load balanced. each service has the following properties:
    - **matcher**: the path to match the request against. if the request path starts with this string, the request will be directed to this service.
    - **name**: the name of the service.
    - **replicas**: the replicas of the service. each replica has the following properties:
        - **url**: the url of the replica.
            - **metadata**: the metadata of the replica, such as weight.

Examples of configuration files can be found in the [examples](https://github.com/Mo-Fatah/mizan/tree/main/examples) directory.

### Running
Mizan can be run using the following command:
```shell
go run main.go -config-path <path to config file>
```

### Running Examples
1- There are example services that can served as a backend services and can be run by:
```shell
make example-service
```
2- Run Mizan with the example configurations in `./examples` by:
```shell
go run main.go -config-path ./examples/config-rr.yml
```
Now, the server is listening on ports `8080`, `8081` and `8082`. You can `curl` these ports from another terminal and mizan will distribute requests to the backend services

3- Since the example services are running in the background, you can kill them with:
```shell
make kill-example
```



## Roadmap
- [x] Multiple Load Balancing Algorithms
- [x] Continuous Health Check 
- [x] Hot Configuration Reloading without Restarting
- [ ] TLS Support
- [ ] Layer 4 Load Balancing
- [ ] HTTP/2 Support
- [ ] Add OpenTelemtry Instrumentation
- [ ] More comprehensive tests