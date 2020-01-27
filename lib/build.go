package lib

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
)

/*
   alpmbuild — a tool to build arch packages from RPM specfiles

   Copyright (C) 2020  Carson Black

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

func ParsePackage(data string) PackageContext {
	lex := PackageContext{}

	lex.Subpackages = make(map[string]PackageContext)

	currentSubpackage := ""

	// Let's parse the Key: Value things first
	for _, line := range strings.Split(strings.TrimSuffix(data, "\n"), "\n") {
		// We expect a ': ' to be present in any Key: Value line
		if strings.Contains(line, ": ") {
			// Split our line by whitespace
			words := strings.Fields(line)

			// Because we need at least two values for a Key: Value PAIR, make
			// sure we have at least two values
			if len(words) < 2 {
				continue
			}

			// Let's worry about our sources first...
			if strings.HasPrefix(strings.ToLower(words[0]), "source") {
				lex.Sources = append(lex.Sources, evalInlineMacros(words[1], lex))
			}

			// And then our patches...
			if strings.HasPrefix(strings.ToLower(words[0]), "patch") {
				lex.Patches = append(lex.Patches, evalInlineMacros(words[1], lex))
			}

			// Now we produce macros based off our package context.

			fields := reflect.TypeOf(lex)

			// If we're currently operating on a subpackage, then we want to write
			// changes to that subpackage.
			if currentSubpackage != "" {
				fields = reflect.TypeOf(lex.Subpackages[currentSubpackage])
			}

			currentPackage := lex
			if currentSubpackage != "" {
				currentPackage = lex.Subpackages[currentSubpackage]
			}

			num := fields.NumField()

			// Loop through the fields of the package context in order to see if any of the annotated key values match the line we're on
			for i := 0; i < num; i++ {
				field := fields.Field(i)
				// These are the single-value keys, such as Name, Version, Release, and Summary
				if strings.ToLower(words[0]) == field.Tag.Get("key") {
					// We assert that packageContext only has string fields here.
					// If it doesn't, our code will break.
					key := reflect.ValueOf(&currentPackage).Elem().FieldByName(field.Name)
					if key.IsValid() {
						key.SetString(evalInlineMacros(strings.TrimSpace(strings.TrimPrefix(line, words[0])), lex))
					}
				}
				// These are the multi-value keys, such as Requires and BuildRequires
				if strings.ToLower(words[0]) == field.Tag.Get("keyArray") {
					// We assume that packageContext only has string array fields here.
					// If it doesn't, our code will break.
					key := reflect.ValueOf(&currentPackage).Elem().FieldByName(field.Name)
					if key.IsValid() {
						itemArray := strings.Split(evalInlineMacros(strings.TrimSpace(strings.TrimPrefix(line, words[0])), lex), " ")

						key.Set(reflect.AppendSlice(key, reflect.ValueOf(itemArray)))
					}
				}
			}

			if currentSubpackage != "" {
				lex.Subpackages[currentSubpackage] = currentPackage
			} else {
				lex = currentPackage
			}
		}

		// We need to be able to handle subpackages
		if strings.Contains(line, "%package") {
			if subpackageName, hasSubpackage := grabFlagFromString(line, "-n", []string{}); hasSubpackage {
				currentSubpackage = lex.Name + "-" + subpackageName
			} else {
				if splitString := strings.Split(line, " "); len(splitString) >= 2 {
					currentSubpackage = lex.Name + "-" + splitString[1]
				} else {
					panic("%package does not have a name!")
				}
			}
			if currentSubpackage != "" {
				lex.Subpackages[currentSubpackage] = PackageContext{
					Name: currentSubpackage,
				}
			}
		}
	}

	currentStage := NoStage

	// Now it's time for the sections.
	for _, line := range strings.Split(strings.TrimSuffix(data, "\n"), "\n") {
		// Let's ignore blank lines, since they're useless.
		if line == "" {
			continue
		}

		// Let's make sure we don't accidentally include directives in
		// commands
		if isStringInSlice(line, otherDirectives) {
			currentStage = NoStage
			continue
		}

		// Now we check to see if we're switching to a new stage
		if strings.HasPrefix(line, "%prep") {
			currentStage = PrepareStage
			continue
		}
		if strings.HasPrefix(line, "%build") {
			currentStage = BuildStage
			continue
		}
		if strings.HasPrefix(line, "%install") {
			currentStage = InstallStage
			continue
		}

		// If we're in a stage, we want to append some commands to our list
		switch currentStage {
		case PrepareStage:
			lex.Commands.Prepare = append(lex.Commands.Prepare, evalInlineMacros(line, lex))
		case BuildStage:
			lex.Commands.Build = append(lex.Commands.Build, evalInlineMacros(line, lex))
		case InstallStage:
			lex.Commands.Install = append(lex.Commands.Install, evalInlineMacros(line, lex))
		}
	}

	fmt.Printf("Package struct:\n%s\n", prettyPrint(lex))

	lex.BuildPackage()

	return lex
}

// Build : Build a specfile, generating an Arch package.
func Build(pathToRecipe string) error {
	data, err := ioutil.ReadFile(pathToRecipe)
	if err != nil {
		return err
	}
	ParsePackage(string(data))
	return nil
}
