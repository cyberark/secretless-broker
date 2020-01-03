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
    stage('Update Submodules') {
        steps {
            sh 'git submodule update --init --recursive'
        }
    }

    stage('Image Build') {
      steps {
        sh './bin/build'
      }
    }

    stage('Scan Secretless') {
      parallel {
        stage('Scan Secretless Image') {
          steps {
            scanAndReport("secretless-broker:latest", "HIGH")
          }
        }

        stage('Scan For Security with Gosec') {
          steps {
            sh "./bin/check_golang_security -s High -c 'Medium' -b ${env.BRANCH_NAME}"
            junit(allowEmptyResults: true, testResults: 'gosec.output')
          }
        }
      }
    }

    stage('Integration Tests') {
      steps { 
        script {
          def directories = sh (
            returnStdout: true,
            // We run the 'find' directive first on all directories with test files, then run a 'find' directive
            // to make sure they also contain start files. We then take the dirname, and basename respectively. 
            script: 
            '''
            find $(find ./test -name test) -name 'start' -exec dirname {} \\; | xargs -n1 basename
            '''
          ).trim().split()

          def integrationSteps = [:]
        
          // Create an integration test stage for each directory we collected previously.
          // We want to be sure to skip any tests, such as keychain tests, that can only be ran manually.
          directories.each { name -> 
            if (name == "keychain") return
            
            def stepName = "Integration: ${name}"

            integrationSteps[stepName] = {
              sh "./bin/run_integration ${name}"
              junit "**/test/**/junit.xml"
            }
          } 

          parallel integrationSteps
        }
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