@Library('jenkins-helpers') _

def label = "gateway-${UUID.randomUUID().toString()}"


podTemplate(
    label: label,
    annotations: [
        podAnnotation(key: "jenkins/build-url", value: env.BUILD_URL ?: ""),
        podAnnotation(key: "jenkins/github-pr-url", value: env.CHANGE_URL ?: ""),
    ],
    containers: [
        containerTemplate(name: 'docker',
            command: '/bin/cat -',
            image: 'docker:17.06.2-ce',
            resourceRequestCpu: '300m',
            resourceRequestMemory: '500Mi',
            resourceLimitCpu: '300m',
            resourceLimitMemory: '500Mi',
            ttyEnabled: true),
        containerTemplate(
            name: 'cloud-sdk',
            image: 'google/cloud-sdk:218.0.0',
            command: 'gcloud beta emulators pubsub start --host-port 0.0.0.0:8085',
            resourceRequestCpu: '300m',
            resourceRequestMemory: '500Mi',
            resourceLimitCpu: '300m',
            resourceLimitMemory: '500Mi',
            ttyEnabled: true
        )
    ],
    volumes: [
        secretVolume(
            secretName: 'jenkins-docker-builder',
            mountPath: '/jenkins-docker-builder',
            readOnly: true
        ),
        hostPathVolume(hostPath: '/var/run/docker.sock', mountPath: '/var/run/docker.sock')
    ]) {
    properties([buildDiscarder(logRotator(daysToKeepStr: '30', numToKeepStr: '20'))])
    node(label) {
        def dockerImageTag = ""

        container('jnlp') {
            stageWithNotify('Checkout') {
                checkout(scm)
                imageRevision = sh(returnStdout: true, script: 'git rev-parse --short HEAD').trim()
                buildDate = sh(returnStdout: true, script: 'date +%Y-%m-%dT%H%M').trim()
                dockerImageTag = "${imageRevision}"
            }
        }
        container('docker') {
            stageWithNotify('Build Docker image') {
                sh("#!/bin/sh -e\n"
                    + "apk add --update-cache make")
                sh("#!/bin/sh -e\n"
                    + "make -j8 TAG=$dockerImageTag")
            }
            if (env.BRANCH_NAME == 'master') {
                stageWithNotify('Push Docker image') {
                    sh('#!/bin/sh -e\n' + 'docker login -u _json_key -p "$(cat /jenkins-docker-builder/credentials.json)" https://eu.gcr.io')
                    sh("make push TAG=$dockerImageTag")
                }
            }
        }
    }
}