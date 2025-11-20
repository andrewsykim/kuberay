package get

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"github.com/ray-project/kuberay/kubectl-plugin/pkg/util/client"
	"github.com/ray-project/kuberay/kubectl-plugin/pkg/util/completion"
)

type GetTokenOptions struct {
	cmdFactory cmdutil.Factory
	ioStreams  *genericclioptions.IOStreams
	namespace  string
	cluster    string
}

func NewGetTokenOptions(cmdFactory cmdutil.Factory, streams genericclioptions.IOStreams) *GetTokenOptions {
	return &GetTokenOptions{
		cmdFactory: cmdFactory,
		ioStreams:  &streams,
	}
}

func NewGetTokenCommand(cmdFactory cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	options := NewGetTokenOptions(cmdFactory, streams)

	cmd := &cobra.Command{
		Use:               "token <cluster-name>",
		Short:             "Get the authentication token for a RayCluster.",
		Long:              `Get the authentication token for a RayCluster. The token is stored in a Secret with the same name as the RayCluster.`,
		SilenceUsage:      true,
		ValidArgsFunction: completion.RayClusterCompletionFunc(cmdFactory),
		Args:              cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := options.Complete(args, cmd); err != nil {
				return err
			}

			k8sClient, err := client.NewClient(cmdFactory)
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}
			return options.Run(cmd.Context(), k8sClient)
		},
	}
	return cmd
}

func (options *GetTokenOptions) Complete(args []string, cmd *cobra.Command) error {
	namespace, err := cmd.Flags().GetString("namespace")
	if err != nil {
		return fmt.Errorf("failed to get namespace: %w", err)
	}
	options.namespace = namespace
	if options.namespace == "" {
		options.namespace = "default"
	}

	options.cluster = args[0]

	return nil
}

func (options *GetTokenOptions) Run(ctx context.Context, k8sClient client.Client) error {
	_, err := k8sClient.RayClient().RayV1().RayClusters(options.namespace).Get(ctx, options.cluster, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get RayCluster %s in namespace %s: %w", options.cluster, options.namespace, err)
	}

	secret, err := k8sClient.KubernetesClient().CoreV1().Secrets(options.namespace).Get(ctx, options.cluster, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get secret %s in namespace %s: %w", options.cluster, options.namespace, err)
	}

	token, ok := secret.Data["auth_token"]
	if !ok {
		return fmt.Errorf("secret %s in namespace %s does not contain 'auth_token'", options.cluster, options.namespace)
	}

	fmt.Fprintln(options.ioStreams.Out, string(token))

	return nil
}
