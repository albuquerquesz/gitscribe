package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/albuquerquesz/gitscribe/internal/config"
	"github.com/albuquerquesz/gitscribe/internal/style"
	"github.com/spf13/cobra"
)

var contextAddCmd = &cobra.Command{
	Use:   "add [contexto]",
	Short: "Adiciona contexto para a AI",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return addContext(args[0])
	},
}

func init() {
	contextCmd.AddCommand(contextAddCmd)
}

func addContext(text string) error {
	// Obter path atual
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
	if len(contexts) >= config.MaxContextsPerPath {
		style.Error(fmt.Sprintf("Limite de %d contextos atingido", config.MaxContextsPerPath))
		style.Info("Use 'gs ctx remove' para remover um contexto existente")
		return fmt.Errorf("limite atingido")
	}

	if err := cm.AddContext(path, text); err != nil {
		style.Error(fmt.Sprintf("Erro ao adicionar contexto: %v", err))
		return err
	}

	style.Success(fmt.Sprintf("Contexto adicionado (%d/%d)", len(contexts)+1, config.MaxContextsPerPath))
	return nil
}
