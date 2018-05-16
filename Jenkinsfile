#!/usr/bin/env groovy

pipeline {
  agent { label 'executor-v2' }

  options {
    timestamps()
    buildDiscarder(logRotator(numToKeepStr: '30'))
  }

  stages {
    stage('Build Linux binaries & Docker images') {
      steps {
        sh './build/build.sh'
      }
    }

    stage('Run tests') {
      steps {
        sh './build/test.sh'

        junit 'test/*.xml'
      }
    }
  }

  post {
    always {
      cleanupAndNotify(currentBuild.currentResult)
    }
  }
}
