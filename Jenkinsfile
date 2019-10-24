#!/usr/bin/env groovy

pipeline {
  agent { label 'executor-v2' }

  options {
    timestamps()
    buildDiscarder(logRotator(numToKeepStr: '30'))
  }

  triggers {
    cron(getDailyCronString())
  }

  stages {
    stage('Image Build') {
      steps {
        sh './bin/build'
      }
    }

    stage('Run Tests') {

      parallel {

        stage('Unit tests') {
          steps {
            sh './bin/test_unit'

            junit 'test/unit-test-output/junit.xml'
            cobertura autoUpdateHealth: true, autoUpdateStability: true, coberturaReportFile: 'coverage.xml', conditionalCoverageTargets: '30, 0, 0', failUnhealthy: true, failUnstable: false, lineCoverageTargets: '30, 0, 0', maxNumberOfBuilds: 0, methodCoverageTargets: '30, 0, 0', onlyStable: false, sourceEncoding: 'ASCII', zoomCoverageChart: false
            ccCoverage("gocov", "--prefix github.com/cyberark/secretless-broker")
          }
        }

        stage('Integration: AWS Secrets Provider') {
          steps {
            sh './bin/run_integration aws_secrets_provider'
            junit 'test/aws_secrets_provider/junit.xml'
          }
        }

        stage('Integration: Conjur') {
          steps {
            sh './bin/run_integration conjur'
            junit 'test/conjur/junit.xml'
          }
        }

        stage('Integration: HTTP Basic Auth') {
          steps {
            sh './bin/run_integration http_basic_auth'
          }
        }

        stage('Integration: Kubernetes Provider') {
          steps {
            sh './bin/run_integration kubernetes_provider'
            junit 'test/kubernetes_provider/junit.xml'
          }
        }

        stage('Integration: MySQL Handler') {
          steps {
            sh './bin/run_integration mysql_handler'
            junit 'test/mysql_handler/junit.xml'
          }
        }

        stage('Integration: MSSQL Handler') {
          steps {
            sh './bin/run_integration mssql_connector'
            junit 'test/mssql_connector/junit.xml'
          }
        }

        stage('Integration: PG Handler') {
          steps {
            sh './bin/run_integration pg_handler'
            junit 'test/pg_handler/junit.xml'
          }
        }

        stage('Integration: SSH Agent Handler') {
          steps {
            sh './bin/run_integration ssh_agent_handler'
            junit 'test/ssh_agent_handler/junit.xml'
          }
        }

        stage('Integration: SSH Handler') {
          steps {
            sh './bin/run_integration ssh_handler'
            junit 'test/ssh_handler/junit.xml'
          }
        }

        stage('Integration: Summon 2') {
          steps {
            sh './bin/run_integration summon2'
            junit 'test/summon2/junit.xml'
          }
        }

        stage('Integration: Vault Provider') {
          steps {
            sh './bin/run_integration vault_provider'
            junit 'test/vault_provider/junit.xml'
          }
        }

        stage('Integration: Template Connector') {
          steps {
            sh './bin/run_integration template_connector'
            junit 'test/template_connector/junit.xml'
          }
        }

        stage('Quick start') {
          steps {
            sh './bin/test_demo'
          }
        }

        stage('K8s Demo') {
          steps {
            sh 'summon -f ./k8s-ci/secrets.yml ./k8s-ci/test demos/k8s-demo'
          }
        }

        stage('CRD test') {
          steps {
            sh 'summon -f ./k8s-ci/secrets.yml ./k8s-ci/test k8s-ci/k8s_crds'
          }
        }
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
            sh 'summon -e production bin/publish_website'
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
