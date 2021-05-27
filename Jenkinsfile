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
    stage('Validate') {
      parallel {
        stage('Changelog') {
          steps { sh './bin/parse-changelog' }
        }
      }
    }

    stage('Update Submodules') {
      steps {
        sh 'git submodule update --init --recursive'
      }
    }

    stage('Build and Unit tests') {
      parallel {
        stage('Build Images') {
          steps {
            sh './bin/build'
          }
        }

        stage('Unit tests') {
          steps {
            sh './bin/test_unit'
            sh 'cp ./test/unit-test-output/c.out ./c.out'

            junit 'test/unit-test-output/junit.xml'
            cobertura autoUpdateHealth: true, autoUpdateStability: true, coberturaReportFile: 'test/unit-test-output/coverage.xml', conditionalCoverageTargets: '30, 0, 0', failUnhealthy: true, failUnstable: false, lineCoverageTargets: '30, 0, 0', maxNumberOfBuilds: 0, methodCoverageTargets: '30, 0, 0', onlyStable: false, sourceEncoding: 'ASCII', zoomCoverageChart: false
            ccCoverage("gocov", "--prefix github.com/cyberark/secretless-broker")
          }
        }
      }
    }

    stage('Scan Secretless') {
      parallel {
        stage('Scan Secretless Image for fixable issues') {
          steps {
            scanAndReport("secretless-broker:latest", "HIGH", false)
          }
        }

        stage('Scan Secretless Image for all issues') {
          steps {
            scanAndReport("secretless-broker:latest", "NONE", true)
          }
        }

        stage('Scan Secretless Quickstart for fixable issues') {
          steps {
            scanAndReport("secretless-broker-quickstart:latest", "HIGH", false)
          }
        }

        stage('Scan Secretless Quickstart for all issues') {
          steps {
            scanAndReport("secretless-broker-quickstart:latest", "NONE", true)
          }
        }

        stage('Scan Secretless RedHat for fixable issues') {
          steps {
            script {
              TAG = sh(returnStdout: true, script: '. bin/build_utils && full_version_tag')
            }
            scanAndReport("secretless-broker-redhat:${TAG}", "HIGH", false)
          }
        }

        stage('Scan Secretless RedHat for all issues') {
          steps {
            script {
              TAG = sh(returnStdout: true, script: '. bin/build_utils && full_version_tag')
            }
            scanAndReport("secretless-broker-redhat:${TAG}", "NONE", true)
          }
        }

        stage('Scan For Security with Gosec') {
          // Gosec only works on branch builds
          when {
            not { tag "v*" }
          }

          steps {
            sh "./bin/check_golang_security -s High -c Medium -b ${env.BRANCH_NAME}"
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

          def integrationStages = [:]

          // Create an integration test stage for each directory we collected previously.
          // We want to be sure to skip any tests, such as keychain tests, that can only be ran manually.
          directories.each { name ->
            if (name == "keychain") return

            integrationStages["Integration: ${name}"] = {
              sh "./bin/run_integration ${name}"
            }
          }

          parallel integrationStages
        }
        junit "**/test/**/junit.xml"
      }
    }

    stage('Functional Tests') {
      parallel {
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

    stage('Push Images Internally') {
      when {
        branch 'main'
      }

      steps {
        sh './bin/publish_internal'
      }
    }

    stage('Build Release Artifacts') {
      when {
        branch 'main'
      }

      steps {
        sh './bin/build_release --snapshot'
        archiveArtifacts 'dist/goreleaser/'
      }
    }

    stage('Release') {
      // Only run this stage when triggered by a tag
      when { tag "v*" }

      parallel {
        stage('Push Images') {
          steps {
            // The tag trigger sets TAG_NAME as an environment variable
            sh 'summon -e production ./bin/publish'
          }
        }
        stage('Create draft release') {
          steps {
            dir('./pristine-checkout') {
              // Go releaser requires a pristine checkout
              checkout scm
              sh 'git submodule update --init --recursive'
              // Create draft release
              sh "summon --yaml 'GITHUB_TOKEN: !var github/users/conjur-jenkins/api-token' ./bin/build_release"
              archiveArtifacts 'dist/goreleaser/'
            }
          }
        }
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
            branch 'main'
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
