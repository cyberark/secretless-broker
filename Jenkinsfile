#!/usr/bin/env groovy

pipeline {
  agent { label 'executor-v2' }

  options {
    timestamps()
    buildDiscarder(logRotator(numToKeepStr: '30'))
  }

  stages {
    stage('Build Binaries & Images') {
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

    stage('Push Images') {
      when {
        branch 'master'
      }

      steps {
        sh './bin/publish'
      }
    }

    stage('Build Website') {
      steps {
        sh './bin/build_website'
      }
    }

    stage('Check Links') {
      steps {
        sh './bin/check_website_links'
      }
    }

    stage('Publish') {
      parallel {
        stage('Publish Website (staging)') {
          when {
            branch 'staging'
          }
          steps {
            sh 'summon -e staging bin/publish_website'
            archiveArtifacts '_site/'
          }
        }

        stage('Publish Website (production)') {
          when {
            branch 'master'
          }
          steps {
            sh 'echo "Skipping production website push - pushing to staging"'
            sh 'summon -e staging bin/publish_website'
            //sh 'summon -e production bin/publish_website'
            archiveArtifacts '_site/'
          }
        }
      }
    }
  }

  post {
    always {
      cleanupAndNotify(currentBuild.currentResult)
    }
  }
}
