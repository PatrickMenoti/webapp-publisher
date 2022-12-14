# webapp-publisher
Publishes your web applications to the Azion Platform.

## Notes

- In order to use this action, you will need go 1.19 in your runner. You may use an image with this go version already installed or use [setup-go](https://github.com/actions/setup-go) action, specifying the desired version. 

- Keep in mind that this action downloads [azioncli](https://www.azion.com/en/documentation/products/CLI) in order to work. Currently we are download the binary for the following OS/architecture: `linux` `x86_64`.



|ENV|DESCRIPTION|OPTIONAL|
|---|---|---|
|`PROJECT_NAME`|The name you want to use to initialize your project|Required|
|`PROJECT_TYPE`|javascript, nextjs, flareact|Required|
|`AZION_TOKEN`|Token used to create the necessary resources in your azion account|Required|
|`AWS_ACCESS_KEY_ID`|Your AWS access key id|Optional - used only if project type is nextjs or flareact|
|`AWS_SECRET_ACCESS_KEY`|Your AWS secret access key|Optional - used only if project type is nextjs or flareact|
|`SETUP_KV`|Indicate true if you wish to setup your own AWS bucket|Optional|
|`KV_BUCKET`|Your AWS Bucket|Optional - used only if SETUP_KV is true|
|`KV_REGION`|Your AWS Region|Optional - used only if project type is nextjs or flareact|
|`KV_PATH`|Your AWS path|Optional - used only if project type is nextjs or flareact|
|`FORCE_INIT`|Indicates if you want to force a new initialization even if you already have an azion template initialized|Optional|
|`SHOULD_COMMIT`|Indicate true if you wish to commit/push changes in `azion` directory to remote|Optional|


## SHOULD_COMMIT

If you do not send `SHOULD_COMMIT` webapp-publisher will not commit the creation and/or changes in `azion` directory; meaning, the next time the action runs, your project will be initialized with a new template. 

-----

## AZIONCLI

You may also download [azioncli](https://www.azion.com/en/documentation/products/CLI) and manually initialize your project and commit those changes to your repository; webapp-publisher will use the template already initialized, if done this way. 