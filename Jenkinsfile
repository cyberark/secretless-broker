#!/usr/bin/env groovy

pipeline {
  agent { label 'executor-v2' }

  options {
    timestamps()
    buildDiscarder(logRotator(numToKeepStr: '30'))
  }

  stages {
    stage('Image Build') {
      steps {
        sh './bin/build'
      }
    }

    stage('Linting') {
      steps {
        sh './bin/check_style'

        checkstyle pattern: 'test/golint.xml', canComputeNew: false, unstableTotalAll: '0', healthy: '0', failedTotalAll: '20',  unHealthy: '10'
      }
    }

    stage('Unit tests') {
      steps {
        sh './bin/test_unit'

        junit 'test/junit.xml'
      }
    }

    stage('Integration tests') {
      steps {
        sh './bin/test_integration'

        junit 'test/junit.xml'
      }
    }

    stage('Demo tests') {
      steps {
        sh './bin/test_quickstart'
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

    stage('Fix Website Flags (staging)') {
      when {
        branch 'staging'
      }
      steps {
        sh 'sed -i "s#^is_maintenance_mode:.*#is_maintenance_mode: false#" docs/_config.yml'
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
            archiveArtifacts 'docs/_site/'
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
            archiveArtifacts 'docs/_site/'
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
