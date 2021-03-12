# Jupyterhub Load Test

This test will execute Tensorflow and Pytorch code in Jupyterhub from the ODS(Open Data Scientist) on OSD(OpenShift Dedicated).

This ODH Operator Test Harness will create a job that will deploy odh manifests test image. The odh manifests test image will create a kfdef for Jupyterhub, then it will create Jupyterhubthe with ODS addon on OSD.

![Image](images/OSD_E2E_Test_Flow_For_Load_Test.png)

## How to test?

- **Pre-requisites**
  ~~~
  git clone https://github.com/Jooho/odh-operator-test-harness.git
  cd  odh-operator-test-harness
  make cluster-test-setup
  ~~~
  
- **For ODH Manifest Test Job** 
  ~~~
  # Update Environment variable of the job.yaml under template folder
  $ vi template/odh-manifest-test-job.yaml

  # Create the Job
  $ oc project redhat-ods-applications
  $ oc create -f template/odh-manifest-test-job.yaml
  $ oc logs job/odh-manifests-test-job -f
 
  ~~~

- **For ODH Test Harness**
  ~~~
  # Update Makefile about the Image registry variable because for test, using your own repository is much easier.
  $vi Makefile
  ...
  DEFAULT_IMAGE_REGISTRY=quay.io
  DEFAULT_REGISTRY_NAMESPACE=jooholee
  DEFAULT_IMAGE_TAG=latest
  ...

  # Build and Push the image
  $ make image

  # Start Test
  $ make cluster-test

  # Clean up test 
  $ make cluster-test-clean
  ~~~