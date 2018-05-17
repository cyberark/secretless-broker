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
        sh './build/build'
      }
    }

    stage('Run tests') {
      steps {
        sh './build/test'

        junit 'test/*.xml'
      }
    }

    stage('Push images') {
      when {
        branch 'master'
      }

      steps {
        sh './build/publish'
      }
    }
  }

  post {
    always {
      cleanupAndNotify(currentBuild.currentResult)
    }
  }
}
