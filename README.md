# odh-operator-test-harness
Test harness for Open Data Hub operator

### Running Locally
To run the test locally, you will need to build the image and run the container image with a valid kubeconfig file mounted in the file

```
$ make
$ podman build -f Dockerfile -t odh-operator-test-harness:latest .
$ podman run --rm -v ~/.kube:/.kube:z -it odh-operator-test-harness:latest

  Running Suite: Odh Operator Test Harness
  ========================================
  Random Seed: 1635881932
  Will run 1 of 1 specs

  •
  JUnit report was created: /test-run-results/junit-odh-operator.xml

  Ran 1 of 1 Specs in 0.119 seconds
  SUCCESS! -- 1 Passed | 0 Failed | 0 Pending | 0 Skipped
  PASS
```
