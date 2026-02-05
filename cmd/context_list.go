package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/albuquerquesz/gitscribe/internal/config"
	"github.com/albuquerquesz/gitscribe/internal/style"
	"github.com/spf13/cobra"
)

var contextListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista contextos do projeto atual",
	RunE: func(cmd *cobra.Command, args []string) error {
		return listContexts()
	},
}

func init() {
	contextCmd.AddCommand(contextListCmd)
}

func listContexts() error {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		style.Error("Não foi possível determinar o diretório do projeto")
		return err
	}
	path := strings.TrimSpace(string(output))

	cm, err := config.LoadContexts()
	if err != nil {
		style.Error(fmt.Sprintf("Erro ao carregar contextos: %v", err))
		return err
	}

	contexts := cm.ListContexts(path)

	fmt.Println(style.TitleStyle.Render(fmt.Sprintf("\n Contextos para %s (%d/%d):", path, len(contexts), config.MaxContextsPerPath)))

	if len(contexts) == 0 {
		style.Info("Nenhum contexto configurado")
		style.Info("Use 'gs ctx add \"seu contexto\"' para adicionar")
		return nil
	}

	for i, ctx := range contexts {
		fmt.Printf("%d. %s\n", i+1, ctx.Text)
	}

	return nil
}
