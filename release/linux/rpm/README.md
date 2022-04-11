RPM Release

## Building RPM in CI/CD pipeline

Execute the following command from the root directory of this repository:

``` bash
./release/linux/rpm/pipeline.sh
```
Output will be sent to `./output/rpm`

To test the packages:

``` bash
./release/linux/rpm/pipeline-test.sh
```


