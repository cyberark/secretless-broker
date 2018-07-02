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

    stage('Run tests') {
      steps {
        sh './bin/test'

        junit 'test/*.xml'
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

    stage('Build site') {
      steps {
        sh './bin/build_website'
        archiveArtifacts '_site/'
      }
    }

    stage('Check links') {
      steps {
        sh './bin/check_website_links'
      }
    }

    stage('Publish') {
      parallel {
        stage('Publish website (staging)') {
          when {
            branch 'staging'
          }
          steps {
            sh 'summon -e staging bin/publish_website'
          }
        }

        stage('Publish website (production)') {
          when {
            branch 'master'
          }
          steps {
            sh 'summon -e production bin/publish_website'
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
