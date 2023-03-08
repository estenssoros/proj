package cmd

import (
	"html/template"
	"os"
	"os/exec"
	"path/filepath"

	_ "embed"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:   "proj",
	Short: "",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("please supply one argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, arg := range args {
			if err := makeProject(arg); err != nil {
				logrus.Errorf("error making project: %s: %s", arg, err.Error())
			}
		}
		return nil
	},
}

func Execute() error {
	return cmd.Execute()
}

func makeProject(projectName string) error {
	logrus.Infof("making project: %s", projectName)
	_, err := os.Stat(projectName)
	if err == nil {
		return errors.Errorf("%s already exists", projectName)
	}
	if err := makeDirectories(projectName); err != nil {
		return errors.Wrap(err, "makeDirectories")
	}

	{
		cmd := exec.Command("go", "mod", "init")
		cmd.Dir = projectName
		logrus.Info(cmd.Args)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return errors.Wrapf(err, "cmd.CombinedOutput: %s", string(out))
		}
	}

	if err := writeTemplates(projectName); err != nil {
		return errors.Wrap(err, "writeTemplates")
	}

	{
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir = projectName
		logrus.Info(cmd.Args)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return errors.Wrapf(err, "cmd.CombinedOutput: %s", string(out))
		}
	}
	logrus.Infof("completed %s", projectName)
	return nil
}

func makeDirectories(projectName string) error {
	err := os.Mkdir(projectName, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "os.Mkdir")
	}
	err = os.Mkdir(filepath.Join(projectName, "cmd"), os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "os.Mkdir")
	}
	return nil
}

//go:embed templates/main.go.tpl
var main_go string

//go:embed templates/cmd.go.tpl
var cmd_go string

func writeTemplates(projectName string) error {
	wd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "os.Getwd")
	}
	var fmt = struct {
		Wd          string
		ProjectName string
	}{
		Wd:          filepath.Base(wd),
		ProjectName: projectName,
	}
	if err := writeTemplateFile(main_go, fmt, projectName, "main.go"); err != nil {
		return errors.Wrap(err, "writeTemplateFile")
	}
	if err := writeTemplateFile(cmd_go, fmt, projectName, "cmd", "cmd.go"); err != nil {
		return errors.Wrap(err, "writeTemplateFile")
	}
	return nil

}

func writeTemplateFile(templateString string, data interface{}, path ...string) error {
	f, err := os.Create(filepath.Join(path...))
	if err != nil {
		return errors.Wrap(err, "os.Create")
	}
	defer f.Close()
	tpl, err := template.New("").Parse(templateString)
	if err != nil {
		return errors.Wrap(err, "template.New.Parse")
	}
	return tpl.Execute(f, data)
}
