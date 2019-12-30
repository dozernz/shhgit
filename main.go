package main

import (
	"bufio"
	"bytes"
	"regexp"
	"strings"

	"github.com/eth0izzle/shhgit/core"
	"github.com/fatih/color"
)

var session = core.GetSession()

func ProcessRepositories() {
	threadNum := *session.Options.Threads

	for i := 0; i < threadNum; i++ {
		go func(tid int) {

			for {
				repositoryId := <-session.Repositories
				repo, err := core.GetRepository(session, repositoryId)

				if err != nil {
					session.Log.Warn("Failed to retrieve repository %d: %s", repositoryId, err)
					continue
				}

				if repo.GetPermissions()["pull"] &&
					uint(repo.GetStargazersCount()) >= *session.Options.MinimumStars &&
					uint(repo.GetSize()) < *session.Options.MaximumRepositorySize {

					processDir(repo.GetCloneURL())
				}
			}
		}(i)
	}
}


func processDir(path string) {
	var (
		matches    []string
	)

	dir := core.GetAbs(path)

	session.Log.Important("[DEBUG] Scanning files in %s", path)

	for _, file := range core.GetMatchingFiles(dir) {
		relativeFileName := strings.Replace(file.Path, *session.Options.TempDirectory, "", -1)
        session.Log.Debug("[DEBUG] Scanning file %s",file.Path)
		if *session.Options.SearchQuery != "" {
			queryRegex := regexp.MustCompile(*session.Options.SearchQuery)
			for _, match := range queryRegex.FindAllSubmatch(file.Contents, -1) {
				matches = append(matches, string(match[0]))
			}

			if matches != nil {
				count := len(matches)
				m := strings.Join(matches, ", ")
				session.Log.Important("[%s] %d %s for %s in file %s: %s", path, count, core.Pluralize(count, "match", "matches"), color.GreenString("Search Query"), relativeFileName, color.YellowString(m))
				session.WriteToCsv([]string{path, "Search Query", relativeFileName, m})
			}
		} else {
			for _, signature := range session.Signatures {
				if matched, part := signature.Match(file); matched {
					//matchedAny = true

					if part == core.PartContents {
						if matches = signature.GetContentsMatches(file); matches != nil {
							count := len(matches)
							m := strings.Join(matches, ", ")
							session.Log.Important("[%s] %d %s for %s in file %s: %s", path, count, core.Pluralize(count, "match", "matches"), color.GreenString(signature.Name()), relativeFileName, color.YellowString(m))
							session.WriteToCsv([]string{path, signature.Name(), relativeFileName, m})
						}
					} else {
						if *session.Options.PathChecks {
							session.Log.Important("[%s] Matching file %s for %s", path, color.YellowString(relativeFileName), color.GreenString(signature.Name()))
							session.WriteToCsv([]string{path, signature.Name(), relativeFileName, ""})
						}

						if *session.Options.EntropyThreshold > 0 && file.CanCheckEntropy() {
							scanner := bufio.NewScanner(bytes.NewReader(file.Contents))

							for scanner.Scan() {
								line := scanner.Text()

								if len(line) > 6 && len(line) < 100 {
									entropy := core.GetEntropy(scanner.Text())

									if entropy >= *session.Options.EntropyThreshold {
										session.Log.Important("[%s] Potential secret in %s = %s", path, color.YellowString(relativeFileName), color.GreenString(scanner.Text()))
										session.WriteToCsv([]string{path, signature.Name(), relativeFileName, scanner.Text()})
									}
								}
							}
						}
					}
				}
			}
		}

		/*if !matchedAny {
			os.Remove(file.Path)
		}*/
	}

	/*if !matchedAny {
		os.RemoveAll(dir)
	}*/
}

func main() {
	session.Log.Info("%s v%s started. Loaded %d signatures. Using %d threads. ", core.Name, core.Version, len(session.Signatures),*session.Options.Threads,)

	if *session.Options.SearchQuery != "" {
		session.Log.Important("Search Query '%s' given. Only returning matching results.", *session.Options.SearchQuery)
	}

    if *session.Options.SkipBinaries == true {
        session.Log.Info("Skipping binaries")
    }


    if *session.Options.TargetDir != "" {
        processDir(*session.Options.TargetDir)
    } else{
        session.Log.Info("Did not select a dir to process")
    }


	//go core.GetRepositories(session)
	//go ProcessRepositories()


	//if *session.Options.ProcessGists {
	//	go core.GetGists(session)
	//	go ProcessGists()
	//}

	session.Log.Info("Completed.\n")
	//select {}
}
