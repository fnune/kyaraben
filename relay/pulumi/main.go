package main

import (
	"github.com/koyeb/pulumi-koyeb/sdk/go/koyeb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		app, err := koyeb.NewKoyebApp(ctx, "kyaraben-relay", &koyeb.KoyebAppArgs{
			Name: pulumi.String("kyaraben-relay"),
		})
		if err != nil {
			return err
		}

		service, err := koyeb.NewKoyebService(ctx, "relay", &koyeb.KoyebServiceArgs{
			AppName: app.Name,
			Definition: &koyeb.KoyebServiceDefinitionArgs{
				Name:    pulumi.String("relay"),
				Regions: pulumi.StringArray{pulumi.String("fra")},
				Docker: &koyeb.KoyebServiceDefinitionDockerArgs{
					Image: pulumi.String("ghcr.io/fnune/kyaraben-relay:latest"),
				},
				InstanceTypes: &koyeb.KoyebServiceDefinitionInstanceTypesArgs{
					Type: pulumi.String("free"),
				},
				Ports: koyeb.KoyebServiceDefinitionPortArray{
					&koyeb.KoyebServiceDefinitionPortArgs{
						Port:     pulumi.Int(8080),
						Protocol: pulumi.String("http"),
					},
				},
				Routes: koyeb.KoyebServiceDefinitionRouteArray{
					&koyeb.KoyebServiceDefinitionRouteArgs{
						Path: pulumi.String("/"),
						Port: pulumi.Int(8080),
					},
				},
				HealthChecks: koyeb.KoyebServiceDefinitionHealthCheckArray{
					&koyeb.KoyebServiceDefinitionHealthCheckArgs{
						Http: &koyeb.KoyebServiceDefinitionHealthCheckHttpArgs{
							Port: pulumi.Int(8080),
							Path: pulumi.String("/health"),
						},
					},
				},
				Scalings: &koyeb.KoyebServiceDefinitionScalingsArgs{
					Min: pulumi.Int(1),
					Max: pulumi.Int(1),
				},
			},
		})
		if err != nil {
			return err
		}

		ctx.Export("appName", app.Name)
		ctx.Export("serviceName", service.ID())

		return nil
	})
}
