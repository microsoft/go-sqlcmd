# Docker Release

## Building Docker in CI/CD pipeline

Execute the following command from the root directory of this repository:

```bash
./release/linux/docker/pipeline.sh
```

Output will be sent to `./output/docker`

## Verify

```bash
./release/linux/docker/pipeline-test.sh
```
