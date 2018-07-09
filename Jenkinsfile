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
        sh './bin/build'
      }
    }

    stage('Static analysis - code linting') {
      steps {
        sh './bin/check_style'

        checkstyle pattern: 'test/golint.xml', canComputeNew: false, unstableTotalAll: '0', healthy: '0', failedTotalAll: '20',  unHealthy: '10'
      }
    }

    stage('Run tests') {
      steps {
        sh './bin/test'

        junit 'test/junit.xml'
      }
    }

    stage('Push images') {
      when {
        branch 'master'
      }

      steps {
        sh './bin/publish'
      }
    }
  }

  post {
    always {
      cleanupAndNotify(currentBuild.currentResult)
    }
  }
}
