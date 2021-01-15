@Library('dst-shared@master') _

dockerBuildPipeline {
        githubPushRepo = "Cray-HPE/hms-sls"
        repository = "cray"
        imagePrefix = "cray"
        app = "sls"
        name = "hms-sls"
        description = "Cray System Layout Service"
        dockerfile = "Dockerfile"
        slackNotification = ["", "", false, false, true, true]
        product = "csm"
}