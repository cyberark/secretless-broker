#!/usr/bin/env groovy

pipeline {
  agent { label 'updated' }

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

    stage('Run Tests') {
      steps {
        sh './bin/test'

        junit 'test/*.xml'
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
            sh 'summon -e production bin/publish_website'
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
