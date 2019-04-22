# basiq-test

This was a task sent from good people at Basiq for a techical review, as a part of an interview.

## Before running

Well, there are few simple things needing to be done before running this:

1. [Sign-up](https://dashboard.basiq.io/login) to the Basiq API service
2. Grab your API key for your application from the [Developer Dashboard](https://dashboard.basiq.io/api-keys))
3. (Optional) In root folder create **config.json** file, paste API key there like this

```json
{
  "APIKey": "YOUR API KEY GOES HERE"
}
```

## Running

To run this small programm there are basically 2 options, and that depends if optional instruction from previous section was completed.

1. If **config.json** was created, just do:

```
go run main.go
```

2. If **config.json** was **not** created, than do:

```
env API_KEY="YOUR API KEY HERE" go run main.go
```
