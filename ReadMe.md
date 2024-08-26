
How to run

1. Run `make run-cast` to simulate Cast environment. There should be no errors printed.
2. In separate shell, run `make run-proxy` to simulate customer environment. There should be no errors printed.
3. In the shell where cast is running, press Return to kick-off the test. Three requests should be made and each should succeed.