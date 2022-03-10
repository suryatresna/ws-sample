docker_build('ws-sample','.', dockerfile='Dockerfile')

k8s_yaml([
    'k8s-ws-sample.yaml',
    'k8s-redis.yaml',
],  allow_duplicates=True)

k8s_resource(
    workload='ws-sample', 
    port_forwards=[
        port_forward(8000, 8000, name='ws-sample-app')
        ],
    resource_deps=["redis-ws-sample"])

k8s_resource(
    workload='redis-ws-sample', 
    port_forwards=[
        port_forward(6379, 6379)
    ])