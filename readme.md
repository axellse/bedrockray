# bedrockray
run a bedrock minecraft server with ray. only linux amd64 supported for now.
## how to use?
it is recommnded to fork the repo so you can edit the config.

specify a unique identifer for the server via the enviroument variable `MCID` and a place to put the server with `MCDIR`. since ray router is http only, dont specify a domain.

you can specify where to download the server from with `MCSERVERDL`