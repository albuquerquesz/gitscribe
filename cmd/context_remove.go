package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/albuquerquesz/gitscribe/internal/config"
	"github.com/albuquerquesz/gitscribe/internal/style"
	"github.com/spf13/cobra"
)

var contextRemoveCmd = &cobra.Command{
	Use:     "remove",
	Short:   "Remove um contexto",
	Aliases: []string{"rm"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return removeContext()
	},
}

func init() {
	contextCmd.AddCommand(contextRemoveCmd)
}

func removeContext() error {
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
	if len(contexts) == 0 {
		style.Info("Nenhum contexto configurado para este projeto")
		return nil
	}

	// Mostrar contextos
	fmt.Println(style.TitleStyle.Render("\n Contextos disponíveis:"))
	for i, ctx := range contexts {
		fmt.Printf("%d. %s\n", i+1, ctx.Text)
	}
	fmt.Println()

	// Perguntar qual remover
	var index int
	fmt.Print("Qual deseja remover? (1-" + fmt.Sprintf("%d", len(contexts)) + "): ")
	if _, err := fmt.Scanf("%d", &index); err != nil || index < 1 || index > len(contexts) {
		style.Info("Operação cancelada")
		return nil
	}

	if err := cm.RemoveContext(path, index-1); err != nil {
		style.Error(fmt.Sprintf("Erro ao remover contexto: %v", err))
		return err
	}

	style.Success("Contexto removido")
	return nil
}
