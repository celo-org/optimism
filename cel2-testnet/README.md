# CEL2 Testnet Setup

Use [direnv](https://direnv.net/) or source `.envrc` to load all required environment variables. [Get the private keys via AKeyless](https://ui.gateway.akeyless.celo-networks-dev.org/items?id=215198970&name=%2Fstatic-secrets%2Fblockchain-circle%2FCel2+testnet+%60.env.secrets%60file+with+wallet+private+keys) and save file in this directory as `.env.secret`.

## Initial Setup

These steps should only be done once when creating a new testnet. The scripts in `initial_setup` are used to create the files in `generated`. Steps:

* Edit .envrc and cel2-testnet.json as necessary.
* Run `initial_setup/deploy_contracts.sh`.
* Run `initial_setup/mk_l2config.sh`.
* Run `initial_setup/mk_jwt_secret.sh`.

## Running the Testnet

This assumes that you run the full testnet setup including the sequencer on our GCP instance. If you want to run a second node, you will have to turn off the sequencer and might have to do further adjustments.

* Copy the `cel2-testnet` directory to the GCP instance `gcloud compute scp --zone "us-west1-b" --project "blockchaintestsglobaltestnet" --recurse ../cel2-testnet/ l2-celo-dev:~`
* Build the docker images with `build_images.sh`.
* Upload the images to the remote docker instance with `upload-images.sh`.
* SSH into the remote host and start the docker-compose setup. `gcloud compute ssh --zone "us-west1-b" "l2-celo-dev" --project "blockchaintestsglobaltestnet" -- 'cd cel2-testnet && ./run_docker.sh'`

## Helper Scripts

The testnet will stop functioning if the accounts run out of balance. Use `balances.sh` to display the current balances.
