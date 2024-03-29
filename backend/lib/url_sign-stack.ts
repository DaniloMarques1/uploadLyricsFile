import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as s3api from 'aws-cdk-lib/aws-s3';
import * as apigateway from 'aws-cdk-lib/aws-apigateway';

export class UrlSignStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);
    const bucketLyrics = new s3api.Bucket(this, "music-lyrics", {
      removalPolicy: cdk.RemovalPolicy.DESTROY,
    });

    const urlSignLambda = new lambda.Function(this, "urlSign", {
      runtime: lambda.Runtime.PROVIDED_AL2023,
      code: lambda.Code.fromAsset("./lambdas"),
      handler: "main",
      environment: {
        BUCKET_NAME: bucketLyrics.bucketName,
      }
    });

    bucketLyrics.grantReadWrite(urlSignLambda);

    const gateway = new apigateway.RestApi(this, "lyrics-api-gateway", {
      defaultCorsPreflightOptions: {
        allowOrigins: ['*'],
        allowMethods: ['PUT'],
      },
      deployOptions: {
        stageName: "dev",
      }
    });

    const integration = new apigateway.LambdaIntegration(urlSignLambda);
    const signResource = gateway.root.addResource("upload");
    signResource.addMethod("PUT", integration);
  }
}
