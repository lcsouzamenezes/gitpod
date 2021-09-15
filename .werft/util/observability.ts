import { werft, exec } from './shell';

/**
 * Monitoring satellite deployment bits
 */
 export class InstallMonitoringSatelliteParams {
    pathToKubeConfig: string
    satelliteNamespace: string
    clusterName: string
    nodeExporterPort: number
}


export async function installMonitoringSatellite(params: InstallMonitoringSatelliteParams) {
    werft.log('observability', `cloning observability repository`)
    exec('git clone https://roboquat:$(cat /mnt/secrets/monitoring-satellite-preview-token/token)@github.com/gitpod-io/observability.git', {silent: true})
    werft.log('observability', 'installing jsonnet utility tools')
    exec('cd observability && make setup-workspace', {silent: true})
    werft.log('observability', 'rendering YAML files')

    // slack_webhook_url_critical and dns_name  needs to be set, but empty, to keep our code
    // compilable and alerting disabled
    // remote_write_urls cannot be empty to make our code compilable
    let jsonnetCmd = `cd observability && jsonnet -c -J vendor -m monitoring-satellite/manifests \
    --ext-str namespace="${params.satelliteNamespace}" \
    --ext-str cluster_name="${params.satelliteNamespace}" \
    --ext-str slack_webhook_url_critical="" \
    --ext-str dns_name="" \
    --ext-code remote_write_urls="['http://fake.endpoint/api/v1/write']" \
    monitoring-satellite/manifests/yaml-generator.jsonnet | xargs -I{} sh -c 'cat {} | gojsontoyaml > {}.yaml' -- {} && \
    find monitoring-satellite/manifests -type f ! -name '*.yaml' ! -name '*.jsonnet'  -delete`

    exec(jsonnetCmd, {silent: true})
    // The correct kubectl context should already be configured prior to this step
    ensureCorrectInstallationOrder()
    exec('sleep 1000000')
}

async function ensureCorrectInstallationOrder(){
    // Adds a label to the namespace metadata
    exec('kubectl apply -f observability/monitoring-satellite/manifests/namespace.yaml', {silent: true})

    exec('kubectl apply -f observability/monitoring-satellite/manifests/podsecuritypolicy-restricted.yaml', {silent: true})
    werft.log('observability', 'installing prometheus-operator')
    exec('kubectl apply -f observability/monitoring-satellite/manifests/prometheus-operator/', {silent: true})
    exec('kubectl rollout status deployment prometheus-operator', {slice: 'observability'})

    deployPrometheus()
    deployGrafana()
    deployNodeExporter()
    deployKubeStateMetrics()
}

async function deployPrometheus() {
    werft.log('observability', 'installing prometheus')
    exec('kubectl apply -f observability/monitoring-satellite/manifests/prometheus/', {silent: true})
    exec('kubectl rollout status statefulset prometheus-k8s', {slice: 'observability'})
}

async function deployGrafana() {
    werft.log('observability', 'installing grafana')
    exec('kubectl apply -f observability/monitoring-satellite/manifests/grafana/', {silent: true})
    // We need to fix https://github.com/gitpod-io/observability/issues/258 first
    // exec('kubectl rollout status deployment grafana', {slice: 'observability'})
}

async function deployNodeExporter() {
    werft.log('observability', 'installing node-exporter')
    exec('kubectl apply -f observability/monitoring-satellite/manifests/node-exporter/', {silent: true})
    exec('kubectl rollout status daemonset node-exporter', {slice: 'observability'})
}

async function deployKubeStateMetrics() {
    werft.log('observability', 'installing kube-state-metrics')
    exec('kubectl apply -f observability/monitoring-satellite/manifests/kube-state-metrics/', {silent: true})
    exec('kubectl rollout status deployment kube-state-metrics', {slice: 'observability'})
}
