import { werft, exec } from './shell';

/**
 * Monitoring satellite deployment bits
 */
 export class InstallMonitoringSatelliteParams {
    pathToKubeConfig: string
    satelliteNamespace: string
    clusterName: string
}


export async function installMonitoringSatellite(params: InstallMonitoringSatelliteParams) {
    werft.log('monitoring-satellite', `cloning observability repository`)
    exec('git clone https://gitpod-gitops-bot:$(cat /mnt/secrets/gitpod-bot-github-token/token)@github.com/gitpod-io/observability')
    exec('cd observability && make setup-workspace')
}