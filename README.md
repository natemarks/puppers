# puppers

This project generates an executable to be used for deployment testing. It only generates logs and the logs include the executable version so we can see the pipeline a little better

The executable just writes a JSON log entries to the '/puppers' cloudwatch log in us-east-1

To build and test locally:

```
make build
./build/current/linux/amd64/puppers
```

## TODO