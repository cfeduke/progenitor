package cmd

import (
  "log"
  "os"
)

import (
  "github.com/caring/progenitor/internal/config"
  "github.com/caring/progenitor/internal/prompt"
  "github.com/caring/progenitor/internal/scaffolding"
  "github.com/caring/progenitor/pkg/aws"
  "github.com/urfave/cli/v2"
)

var (
  awsClient *aws.Client
  cfg       *config.Config
)

var prompts = map[string][]func(*config.Config) error{
  "go-grpc": {
    prompt.ProjectName,
    prompt.ProjectDir,
    prompt.UseDB,
    prompt.CoreDBObject,
    prompt.UseReporting,
    prompt.RunTerraform,
  },
}

func Execute() {

  cfg = config.New()

  app := &cli.App{
    Name: "progenitor",
    Usage: `
             @@@@,             
           (@@@@@@@           
 ,##%.     (@@@@@@@,     *###    Hello, I am the Progenitor!!!
  #####*     @@@@@    *#####* 
    ######          *####*       Please answer my questions, and
       ####*       ####*         I will set up a nice set of 
        .####    .####           boilerplate code, so that you 
          ####  .###*            do not need to do that awful
          .###  ####             copy pasta you used to do.
           '###*###           
           .### ###           
           ###* ###.
          .###  ####          
          ###,   ###,          `,
    Commands: []*cli.Command{
      {
        Name:  "go-grpc",
        Usage: "scaffolds a gRPC service in Go",
        Action: func(c *cli.Context) error {
          cfg.Set("projectType", "go-grpc")
          return generate(cfg)
        },
      },
    },
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Println(err)
  }

}

func generate(cfg *config.Config) error {

  awsClient = setupAwsClient()

  for _, p := range prompts[cfg.GetString("projectType")] {
    if err := p(cfg); err != nil {
      return handleError(err)
    }
  }

  token, err := awsClient.GetSecret("github_token")
  if err != nil {
    return handleError(err)
  }

  _, err = createRepo(*token.SecretString, cfg)
  if err != nil {
    log.Println(err.Error())
    return err
  }

  scaffold, err := scaffolding.New(cfg)
  if err != nil {
    log.Println(err.Error())
    return err
  }

  if err = scaffold.BuildStructure(); err != nil {
    log.Println(err.Error())
    return err
  }

  if err = scaffold.BuildFiles(*token.SecretString); err != nil {
    log.Println(err.Error())
    return err
  }

  if err = commitCodeToRepo(*token.SecretString, scaffold); err != nil {
    log.Println(err.Error())
    return err
  }

  // moved to here, note that this assumes that all projects will store
  // the terraform code in a /terraform folder in the root directory
  // of the project... while this appears true at this time, it may not
  // be in the future. This change enables committing code to the repo
  // independent of the success of terraform running... which was previously
  // breaing the code. There is probably a better long term fix, which we can
  // inverst in if it continues to create issues
  if scaffold.Config.GetBool("runTerraform") {
    base, err := os.Getwd()
    tfDir := filepath.Join(base, scaffold.Config.GetString("projectDir"), "terraform")

    if err := terraform.Run(tfDir); err != nil {
      return err
    }
  }

  return nil
}

func setupAwsClient() *aws.Client {
  var region string = "us-east-1"
  var account_id string = "182565773517"
  var role string = "ops-mgmt-admin"

  awsClient := aws.New()
  awsClient.SetConfig(&region, &account_id, &role)

  return awsClient
}

func handleError(err error) error {
  log.Println(err)
  os.Exit(1)
  return err
}
