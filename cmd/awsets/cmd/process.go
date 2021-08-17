package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/emicklei/dot"
	"github.com/trek10inc/awsets/resource"
	"github.com/urfave/cli/v2"
)

var processCmd = &cli.Command{
	Name:      "process",
	Usage:     "runs processors on results json",
	ArgsUsage: "[input file]",
	Subcommands: []*cli.Command{
		dotGenerator,
		stats,
		cfn,
	},
}

var stats = &cli.Command{
	Name:  "stats",
	Usage: "generates statistics for input file",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:      "input",
			Aliases:   []string{"i"},
			Value:     "",
			Usage:     "input file containing data to process",
			TakesFile: true,
		},
	},
	Action: func(c *cli.Context) error {

		resources, err := loadData(c.String("input"))
		if err != nil {
			return fmt.Errorf("failed to load data: %w", err)
		}

		type regionType struct {
			region string
			kind   string
		}

		counts := make(map[regionType]int)

		for _, res := range resources {
			counts[regionType{
				region: res.Region,
				kind:   res.Type.String(),
			}]++
		}

		for k, v := range counts {
			fmt.Printf("%s,%s,%d\n", k.region, k.kind, v)
		}

		return nil
	},
}

var dotGenerator = &cli.Command{
	Name:  "dot",
	Usage: "generates dot file of relationships in input file",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:      "input",
			Aliases:   []string{"i"},
			Value:     "",
			Usage:     "input file containing data to process",
			TakesFile: true,
		},
		&cli.StringFlag{
			Name:      "output",
			Aliases:   []string{"o"},
			Value:     "",
			Usage:     "output file to save results",
			TakesFile: true,
		},
		&cli.BoolFlag{
			Name:  "show-all",
			Value: false,
			Usage: "include all unrelated items",
		},
	},
	Action: func(c *cli.Context) error {

		resources, err := loadData(c.String("input"))
		if err != nil {
			return fmt.Errorf("failed to load data: %w", err)
		}

		// generate map of all items that have relationships
		nodeIds := make(map[IdLite]string)
		idCounter := 0
		relatedItems := make(map[IdLite]resource.Resource)
		phantomItems := make(map[IdLite]struct{})
		for k, r := range resources {
			nodeIds[k] = fmt.Sprintf("%d", idCounter)
			idCounter++

			if len(r.Relations) > 0 {
				id := genId(r.Identifier)
				relatedItems[id] = resources[id]

				for _, rel := range r.Relations {
					relId := genId(rel)
					if v, ok := resources[relId]; ok {
						relatedItems[relId] = v
					} else {
						phantomItems[relId] = struct{}{}
						nodeIds[relId] = fmt.Sprintf("%d", idCounter)
						idCounter++
					}
				}
			}
		}

		for k := range relatedItems {
			delete(resources, k)
		}

		subgraphs := make(map[string]*dot.Graph)

		graph := dot.NewGraph(dot.Directed)
		graph.Attr("rankdir", "LR")

		// write all nodes from related items
		for k, r := range relatedItems {
			regionGraph, ok := subgraphs[k.Region]
			if !ok {
				regionGraph = graph.Subgraph(k.Region, dot.ClusterOption{})
				subgraphs[k.Region] = regionGraph
			}
			regionGraph.Node(nodeIds[k]).Box().Label(makeLabel(r))
		}
		// write all nodes for phantom items
		for k := range phantomItems {
			regionGraph, ok := subgraphs[k.Region]
			if !ok {
				regionGraph = graph.Subgraph(k.Region, dot.ClusterOption{})
				subgraphs[k.Region] = regionGraph
			}
			regionGraph.Node(nodeIds[k]).Box().Label(makeLabel(k)).Attr("style", "filled").Attr("color", "#FF9898")
		}

		// write all edges for related items
		for k, r := range relatedItems {
			fromGraph := subgraphs[k.Region]
			fromNode, found := fromGraph.FindNodeById(nodeIds[k])
			if !found {
				panic("failed to find 'from' node")
			}
			for _, rel := range r.Relations {
				relId := genId(rel)
				toGraph := subgraphs[relId.Region]
				toNode, found := toGraph.FindNodeById(nodeIds[relId])
				if !found {
					log.Printf("from node %+v, failed to find 'to' node %+v\n", k, relId)
					continue
				}
				fromNode.Edge(toNode)
			}
		}

		// write unrelated items
		for k, r := range resources {
			if c.Bool("show-all") || !isAWSDefault(r) {
				regionGraph, ok := subgraphs[k.Region]
				if !ok {
					regionGraph = graph.Subgraph(k.Region, dot.ClusterOption{})
					subgraphs[k.Region] = regionGraph
				}
				unrelatedGraph, ok := subgraphs[k.Region+"_unrelated"]
				if !ok {
					unrelatedGraph = regionGraph.Subgraph(k.Region+"_unrelated", dot.ClusterOption{})
					unrelatedGraph.Attr("style", "filled")
					unrelatedGraph.Attr("color", "#ffffee")
					subgraphs[k.Region+"_unrelated"] = unrelatedGraph
				}
				unrelatedGraph.Node(nodeIds[k]).Box().Label(makeLabel(r))
			}
		}

		if c.String("output") == "" {
			fmt.Printf("%s", graph.String())
		} else {
			err = ioutil.WriteFile(c.String("output"), []byte(graph.String()), 0655)
			if err != nil {
				log.Fatalf("failed to write output file: %v\n", err)
			}
		}
		return nil
	},
}

var cfn = &cli.Command{
	Name:  "cfn",
	Usage: "filters out resources managed by cloudformation",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:      "input",
			Aliases:   []string{"i"},
			Value:     "",
			Usage:     "input file containing data to process",
			TakesFile: true,
		},
		&cli.StringFlag{
			Name:      "output",
			Aliases:   []string{"o"},
			Value:     "",
			Usage:     "output file to save results",
			TakesFile: true,
		},
	},
	Action: func(c *cli.Context) error {

		resources, err := loadData(c.String("input"))
		if err != nil {
			return fmt.Errorf("failed to load data: %w", err)
		}

		toRemove := make([]IdLite, 0)
		for _, res := range resources {
			if res.Type == resource.CloudFormationStack {
				toRemove = append(toRemove, genId(res.Identifier))
				for _, rel := range res.Relations {
					toRemove = append(toRemove, genId(rel))
				}
			} else if _, hasTag := res.Tags["aws:cloudformation:stack-id"]; hasTag {
				toRemove = append(toRemove, genId(res.Identifier))
			} else if res.Type == resource.Route53RecordSet ||
				res.Type == resource.Route53HostedZone ||
				res.Type == resource.Route53HealthCheck {
				toRemove = append(toRemove, genId(res.Identifier))
			}
		}

		fmt.Printf("pre filtered count: %d\n", len(resources))
		for _, remove := range toRemove {
			delete(resources, remove)
		}
		fmt.Printf("post filtered count: %d\n", len(resources))

		rg := resource.NewGroup()
		for _, v := range resources {
			rg.AddResource(v)
		}
		data, err := rg.JSON()
		if err != nil {
			panic(err)
		}

		if c.String("output") == "" {
			fmt.Printf("%s", data)
		} else {
			err = ioutil.WriteFile(c.String("output"), []byte(data), 0655)
			if err != nil {
				log.Fatalf("failed to write output file: %v\n", err)
			}
		}
		return nil
	},
}

func isAWSDefault(r resource.Resource) bool {
	switch r.Type {
	case resource.CodeDeployDeploymentConfig:
		return strings.Contains(r.Name, "Default.")
	case resource.DAXParameterGroup:
		return strings.HasPrefix(r.Id, "default.")
	case resource.DocDBParameterGroup:
		return strings.HasPrefix(r.Id, "default.")
	case resource.ElasticacheParameterGroup:
		return strings.HasPrefix(r.Id, "default.")
	case resource.IamPolicy:
		exclude := []string{"Alexa", "Amazon", "APIGateway", "AutoScaling", "AWS", "Aws", "CloudFront", "CloudSearch",
			"CloudWatch", "DAX", "EC2", "Ec2", "Elastic", "IAM", "LakeFormation", "Lex", "Neptune", "Translate", "WAF"}
		for _, ex := range exclude {
			if strings.HasPrefix(r.Name, ex) {
				return true
			}
		}
	case resource.IamRole:
		exclude := []string{"Amazon", "AWS"}
		for _, ex := range exclude {
			if strings.HasPrefix(r.Name, ex) {
				return true
			}
		}
	case resource.KmsAlias:
		return strings.HasPrefix(r.Name, "alias/aws/")
	case resource.NeptuneDbParameterGroup:
		return strings.HasPrefix(r.Id, "default.")
	case resource.NeptuneDbClusterParameterGroup:
		return strings.HasPrefix(r.Id, "default.")
	case resource.RdsDbParameterGroup:
		return strings.HasPrefix(r.Id, "default")
	case resource.RdsDbClusterParameterGroup:
		return strings.HasPrefix(r.Id, "default")
	case resource.SsmPatchBaseline:
		return strings.Contains(r.Name, "Default") || strings.Contains(r.Name, "WindowsPredefined")
	}
	return false
}

func makeLabel(i interface{}) string {
	parts := make([]string, 0)
	if v, ok := i.(resource.Resource); ok {
		parts = append(parts, v.Id)
		if v.Name != "" && v.Name != v.Id {
			parts = append(parts, v.Name)
		}
		parts = append(parts, v.Type.String())
	} else if v, ok := i.(IdLite); ok {
		parts = append(parts, v.Id)
		parts = append(parts, v.Type.String())
	}
	return strings.Join(parts, "\n")
}
