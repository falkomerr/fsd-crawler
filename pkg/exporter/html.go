package exporter

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"fsd-crawler/pkg/config"
	"fsd-crawler/pkg/dependencies"
	"fsd-crawler/pkg/model"
)

func GenerateHTML(structure *model.ProjectStructure, cfg *config.Config) error {
	var tmplContent string
	var err error

	if cfg != nil && cfg.HTMLTemplatePath != "" {
		tmplBytes, err := os.ReadFile(cfg.HTMLTemplatePath)
		if err != nil {
			return fmt.Errorf("не удалось прочитать пользовательский HTML шаблон: %v", err)
		}
		tmplContent = string(tmplBytes)
	} else {
		tmplContent = defaultHTMLTemplate
	}

	outputDir := "./dist"
	if cfg != nil && cfg.OutputDir != "" {
		outputDir = cfg.OutputDir
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("не удалось создать директорию для вывода: %v", err)
	}

	outputPath := filepath.Join(outputDir, "fsd_structure.html")
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("не удалось создать HTML файл: %v", err)
	}
	defer outputFile.Close()

	templateData := struct {
		Layers                    []*model.FSDLayer
		Dependencies              []dependencies.Dependency
		HasDependencies           bool
		AllowedCyclicalDependencies []string
	}{
		Layers:                    structure.Layers,
		Dependencies:              []dependencies.Dependency{},
		HasDependencies:           len(structure.Dependencies) > 0,
		AllowedCyclicalDependencies: cfg.AllowedCyclicalDependencies,
	}

	for _, dep := range structure.Dependencies {
		if d, ok := dep.(dependencies.Dependency); ok {
			templateData.Dependencies = append(templateData.Dependencies, d)
		}
	}

	t, err := template.New("fsdStructure").Parse(tmplContent)
	if err != nil {
		return fmt.Errorf("ошибка при парсинге HTML шаблона: %v", err)
	}

	if err := t.Execute(outputFile, templateData); err != nil {
		return fmt.Errorf("ошибка при генерации HTML: %v", err)
	}

	return nil
}

const defaultHTMLTemplate = `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>FSD Structure Analyzer</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            margin: 0;
            padding: 20px;
            color: #333;
        }
        h1, h2 {
            color: #2c3e50;
            margin-bottom: 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        .layer {
            margin-bottom: 30px;
            border: 1px solid #ddd;
            border-radius: 4px;
            overflow: hidden;
        }
        .layer-header {
            background-color: #f5f5f5;
            padding: 10px 15px;
            font-weight: bold;
            border-bottom: 1px solid #ddd;
        }
        .slice {
            margin: 15px;
            border: 1px solid #e0e0e0;
            border-radius: 4px;
            overflow: hidden;
        }
        .slice-header {
            background-color: #f9f9f9;
            padding: 8px 12px;
            font-weight: bold;
            border-bottom: 1px solid #e0e0e0;
        }
        .segment {
            margin: 10px;
            border: 1px solid #eee;
            border-radius: 4px;
        }
        .segment-header {
            background-color: #fafafa;
            padding: 6px 10px;
            font-weight: bold;
            border-bottom: 1px solid #eee;
        }
        .files {
            padding: 5px 10px;
        }
        .file {
            padding: 3px 0;
            font-size: 14px;
        }
        .empty-message {
            padding: 10px;
            color: #999;
            font-style: italic;
        }
        .dependencies-section {
            margin-top: 40px;
        }
        .dependency-list {
            margin-top: 20px;
        }
        .dependency-item {
            padding: 8px;
            margin-bottom: 5px;
            border-radius: 4px;
        }
        .dependency-normal {
            background-color: #d4edda;
            border: 1px solid #c3e6cb;
            color: #155724;
        }
        .dependency-same {
            background-color: #fff3cd;
            border: 1px solid #ffeeba;
            color: #856404;
        }
        .dependency-cyclical {
            background-color: #f8d7da;
            border: 1px solid #f5c6cb;
            color: #721c24;
        }
        .dependency-test {
            background-color: #e2e3e5;
            border: 1px solid #d6d8db;
            color: #383d41;
        }
        .dependency-allowed-cyclical {
            background-color: #d4edda;
            border: 1px solid #c3e6cb;
            color: #155724;
            opacity: 0.7;
        }
        .dependency-graph {
            position: relative;
            height: 600px;
            border: 1px solid #ddd;
            border-radius: 4px;
            margin-top: 20px;
            overflow: hidden;
        }
        .dependency-graph svg {
            width: 100%;
            height: auto;
            min-height: 500px;
            border: 1px solid #ddd;
        }
        .node {
            cursor: pointer;
        }
        .link {
            stroke-width: 2px;
        }
        .link-normal {
            stroke: #28a745;
        }
        .link-same {
            stroke: #ffc107;
        }
        .link-cyclical {
            stroke: #dc3545;
        }
        .link-test {
            stroke: #6c757d;
        }
        .controls {
            position: absolute;
            top: 10px;
            right: 10px;
            background: rgba(255, 255, 255, 0.8);
            padding: 10px;
            border-radius: 4px;
            border: 1px solid #ddd;
            z-index: 10;
        }
        .controls button {
            margin: 0 5px;
            padding: 5px 10px;
            background: #4a69bd;
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
        }
        .controls button:hover {
            background: #3a5795;
        }
        .node-label {
            font-size: 12px;
            pointer-events: none;
        }
        .node circle {
            stroke: #fff;
            stroke-width: 2px;
        }
        .tooltip {
            position: absolute;
            background: white;
            border: 1px solid #ddd;
            border-radius: 4px;
            padding: 10px;
            pointer-events: none;
            opacity: 0;
            transition: opacity 0.3s;
        }
    </style>
    <script src="https://d3js.org/d3.v7.min.js"></script>
</head>
<body>
    <div class="container">
        <h1>Feature-Sliced Design Structure Analyzer</h1>
        
        {{if .Layers}}
            {{range .Layers}}
                <div class="layer">
                    <div class="layer-header">Слой: {{.Name}}</div>
                    
                    {{range .Slices}}
                        <div class="slice">
                            {{if .Name}}
                                <div class="slice-header">Слайс: {{.Name}}</div>
                            {{end}}
                            
                            {{range .Segments}}
                                <div class="segment">
                                    <div class="segment-header">Сегмент: {{.Name}}</div>
                                    {{if .Files}}
                                        <div class="files">
                                            {{range .Files}}
                                                <div class="file">{{.}}</div>
                                            {{end}}
                                        </div>
                                    {{else}}
                                        <div class="files">Нет файлов</div>
                                    {{end}}
                                </div>
                            {{else}}
                                <div class="empty-message">Нет сегментов</div>
                            {{end}}
                        </div>
                    {{else}}
                        <div class="empty-message">Нет слайсов</div>
                    {{end}}
                </div>
            {{end}}
        {{else}}
            <div class="empty-message">FSD структура не обнаружена</div>
        {{end}}

        {{if .HasDependencies}}
            <div class="dependencies-section">
                <h2>Зависимости между слоями и слайсами</h2>
                
                {{if .AllowedCyclicalDependencies}}
                    <div class="allowed-cyclical-info">
                        <h3>Слайсы с разрешенными циклическими зависимостями:</h3>
                        <ul>
                            {{range .AllowedCyclicalDependencies}}
                                <li>{{.}}</li>
                            {{end}}
                        </ul>
                    </div>
                {{end}}
                
                <div class="dependency-graph" id="dependency-graph">
                    <div class="controls">
                        <button id="zoom-in">+</button>
                        <button id="zoom-out">-</button>
                    </div>
                </div>
                
                <div class="dependency-list">
                    <h3>Список зависимостей</h3>
                    {{range .Dependencies}}
                        {{$dep := .}}
                        {{$isAllowedCyclical := false}}
                        {{$fromPath := (print $dep.FromLayer "/" $dep.FromSlice)}}
                        {{$toPath := (print $dep.ToLayer "/" $dep.ToSlice)}}
                        
                        {{if eq $dep.Type "cyclical"}}
                            {{range $index, $element := $.AllowedCyclicalDependencies}}
                                {{if or (eq $element $fromPath) (eq $element $toPath) (eq $element $dep.FromLayer) (eq $element $dep.ToLayer)}}
                                    {{$isAllowedCyclical = true}}
                                {{end}}
                            {{end}}
                        {{end}}
                        
                        <div class="dependency-item {{if $isAllowedCyclical}}dependency-allowed-cyclical{{else}}dependency-{{$dep.Type}}{{end}}">
                            {{$dep.FromLayer}}/{{$dep.FromSlice}} → {{$dep.ToLayer}}/{{$dep.ToSlice}}
                            {{if $isAllowedCyclical}}
                                (разрешенная циклическая зависимость)
                            {{else if eq $dep.Type "normal"}}
                                (нормальная зависимость)
                            {{else if eq $dep.Type "same"}}
                                (зависимость на том же слое)
                            {{else if eq $dep.Type "cyclical"}}
                                (циклическая зависимость)
                            {{else if eq $dep.Type "test"}}
                                (тестовая зависимость)
                            {{end}}
                        </div>
                    {{end}}
                </div>
            </div>

            <script>
                document.addEventListener('DOMContentLoaded', function() {
                    const allowedCyclicalPaths = [
                        {{range .AllowedCyclicalDependencies}}
                            "{{.}}",
                        {{end}}
                    ];
                    
                    const dependencies = [
                        {{range .Dependencies}}
                            {
                                source: "{{.FromLayer}}/{{.FromSlice}}",
                                target: "{{.ToLayer}}/{{.ToSlice}}",
                                type: "{{.Type}}",
                                fromLayer: "{{.FromLayer}}",
                                fromSlice: "{{.FromSlice}}",
                                toLayer: "{{.ToLayer}}",
                                toSlice: "{{.ToSlice}}"
                            },
                        {{end}}
                    ];
                    
                    dependencies.forEach(d => {
                        const fromPath = d.fromLayer + "/" + d.fromSlice;
                        const toPath = d.toLayer + "/" + d.toSlice;
                        
                        if (d.type === "cyclical" && 
                            (allowedCyclicalPaths.includes(fromPath) || 
                             allowedCyclicalPaths.includes(toPath) ||
                             allowedCyclicalPaths.includes(d.fromLayer) ||
                             allowedCyclicalPaths.includes(d.toLayer))) {
                            d.type = "normal";
                            d.isAllowedCyclical = true;
                        }
                    });
                    
                    const nodes = new Set();
                    dependencies.forEach(d => {
                        nodes.add(d.source);
                        nodes.add(d.target);
                    });
                    
                    const nodesArray = Array.from(nodes).map(id => {
                        const parts = id.split('/');
                        const layerName = parts[0];
                        return { 
                            id,
                            layerName,
                            isAllowedCyclical: allowedCyclicalPaths.includes(id) || allowedCyclicalPaths.includes(layerName)
                        };
                    });
                    
                    const links = dependencies.map(d => ({
                        source: d.source,
                        target: d.target,
                        type: d.type,
                        isAllowedCyclical: d.isAllowedCyclical
                    }));
                    
                    const width = document.getElementById('dependency-graph').clientWidth;
                    const height = 600;
                    
                    const svg = d3.select('#dependency-graph')
                        .append('svg')
                        .attr('width', width)
                        .attr('height', height)
                        .attr('viewBox', [0, 0, width, height])
                        .call(d3.zoom()
                            .extent([[0, 0], [width, height]])
                            .scaleExtent([0.1, 4])
                            .on("zoom", zoomed));
                    
                    const tooltip = d3.select('#dependency-graph')
                        .append('div')
                        .attr('class', 'tooltip');
                        
                    const g = svg.append('g');
                    
                    function zoomed(event) {
                        g.attr('transform', event.transform);
                    }
                    
                    const linksGroup = g.append('g').attr('class', 'links');
                    const nodesGroup = g.append('g').attr('class', 'nodes');
                    
                    const simulation = d3.forceSimulation(nodesArray)
                        .force('link', d3.forceLink(links)
                            .id(d => d.id)
                            .distance(100)
                            .strength(0.5))
                        .force('charge', d3.forceManyBody()
                            .strength(-300)
                            .distanceMax(500))
                        .force('center', d3.forceCenter(width / 2, height / 2))
                        .force('collide', d3.forceCollide().radius(50))
                        .force('x', d3.forceX(width / 2).strength(0.05))
                        .force('y', d3.forceY(height / 2).strength(0.05));
                    
                    svg.append('defs').selectAll('marker')
                        .data(['normal', 'same', 'cyclical', 'test', 'allowed-cyclical'])
                        .enter()
                        .append('marker')
                        .attr('id', d => 'arrow-' + d)
                        .attr('viewBox', '0 -5 10 10')
                        .attr('refX', 25)
                        .attr('refY', 0)
                        .attr('markerWidth', 6)
                        .attr('markerHeight', 6)
                        .attr('orient', 'auto')
                        .append('path')
                        .attr('d', 'M0,-5L10,0L0,5')
                        .attr('fill', d => {
                            switch(d) {
                                case 'normal': return '#28a745';
                                case 'same': return '#ffc107';
                                case 'cyclical': return '#dc3545';
                                case 'test': return '#6c757d';
                                case 'allowed-cyclical': return '#28a745';
                                default: return '#28a745';
                            }
                        });
                    
                    const link = linksGroup.selectAll('line')
                        .data(links)
                        .enter()
                        .append('line')
                        .attr('class', d => 'link link-' + d.type)
                        .attr('stroke', d => {
                            if (d.isAllowedCyclical) return '#28a745';
                            switch(d.type) {
                                case 'normal': return '#28a745';
                                case 'same': return '#ffc107';
                                case 'cyclical': return '#dc3545';
                                case 'test': return '#6c757d';
                                default: return '#28a745';
                            }
                        })
                        .attr('stroke-opacity', d => d.isAllowedCyclical ? 0.7 : 1)
                        .attr('marker-end', d => d.isAllowedCyclical ? 
                            'url(#arrow-allowed-cyclical)' : 'url(#arrow-' + d.type + ')');
                    
                    const node = nodesGroup.selectAll('.node')
                        .data(nodesArray)
                        .enter()
                        .append('g')
                        .attr('class', 'node')
                        .call(d3.drag()
                            .on('start', dragstarted)
                            .on('drag', dragged)
                            .on('end', dragended))
                        .on('mouseover', function(event, d) {
                            tooltip.style('opacity', 1)
                                .html('<strong>' + d.id + '</strong>' + 
                                    (d.isAllowedCyclical ? ' (разрешены циклические зависимости)' : ''))
                                .style('left', (event.pageX - document.getElementById('dependency-graph').offsetLeft + 10) + 'px')
                                .style('top', (event.pageY - document.getElementById('dependency-graph').offsetTop - 30) + 'px');
                        })
                        .on('mouseout', function() {
                            tooltip.style('opacity', 0);
                        });
                    
                    node.append('circle')
                        .attr('r', 15)
                        .attr('fill', d => {
                            const color = getNodeColor(d.id);
                            return d.isAllowedCyclical ? d3.color(color).brighter(0.3) : color;
                        })
                        .attr('opacity', d => d.isAllowedCyclical ? 0.8 : 1);
                    
                    function getNodeColor(id) {
                        const layer = id.split('/')[0];
                        switch(layer) {
                            case 'app': return '#3498db';
                            case 'processes': return '#9b59b6';
                            case 'pages': return '#2ecc71';
                            case 'widgets': return '#f1c40f';
                            case 'features': return '#e67e22';
                            case 'entities': return '#e74c3c';
                            case 'shared': return '#95a5a6';
                            default: return '#34495e';
                        }
                    }
                    
                    node.append('text')
                        .attr('class', 'node-label')
                        .attr('dx', 12)
                        .attr('dy', '.35em')
                        .text(d => d.id);
                    
                    simulation.on('tick', () => {
                        link
                            .attr('x1', d => d.source.x)
                            .attr('y1', d => d.source.y)
                            .attr('x2', d => d.target.x)
                            .attr('y2', d => d.target.y);
                        
                        node.attr('transform', d => 'translate(' + d.x + ',' + d.y + ')');
                    });
                    
                    function dragstarted(event, d) {
                        if (!event.active) simulation.alphaTarget(0.3).restart();
                        d.fx = d.x;
                        d.fy = d.y;
                    }
                    
                    function dragged(event, d) {
                        d.fx = event.x;
                        d.fy = event.y;
                    }
                    
                    function dragended(event, d) {
                        if (!event.active) simulation.alphaTarget(0);
                        d.fx = null;
                        d.fy = null;
                    }
                    
                    document.getElementById('zoom-in').addEventListener('click', function() {
                        svg.transition().call(
                            d3.zoom().on('zoom', zoomed).transform,
                            d3.zoomIdentity.scale(d3.zoomTransform(svg.node()).k * 1.3)
                        );
                    });
                    
                    document.getElementById('zoom-out').addEventListener('click', function() {
                        svg.transition().call(
                            d3.zoom().on('zoom', zoomed).transform,
                            d3.zoomIdentity.scale(d3.zoomTransform(svg.node()).k / 1.3)
                        );
                    });
                    
                    simulation.alpha(1).restart();
                });
            </script>
        {{else}}
            <div class="empty-message">Зависимости не обнаружены</div>
        {{end}}
    </div>
</body>
</html>` 