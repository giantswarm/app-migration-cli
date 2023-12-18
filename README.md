# app-migration-cli

This tool migrates `apps` from one MC to another

## :airplane: Rundown

Phases of the script should run in advance of the CAPI Migration, others
are scheduled after a successfull infrastructure migration.

* preflight - *readonly checks if a migration is possible*
    * validate access to both mcs
    * check WC condition/health (only "Created" is allowed)

* pre - *writing all resources to disk*
    * putting a finalizier on the NS to prevent its deletion
    * writing all `apps` to disk
    * writing all depend cm/secrets to disk

* post - *applying the resources to the new MC*
    * applying the dumped resources to the new MC
    * removing the ns/finalizier

## Recomendation to run the tool
To ensure there are no interference with kubeconfigs that the tool uses, create a new temporary file for kubeconfig.
```
export KUBECONFIG=$(mktemp)
chmod 600 $KUBECONFIG
```

## Notes 
* currently only working for vintage
* currently only working in aws
