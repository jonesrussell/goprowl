package metrics

import (
	"html/template"
	"net/http"
)

// Dashboard template with Tailwind CSS 3 and dark mode support
const dashboardTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>GoProwl Metrics Dashboard</title>
    <script src="https://cdn.plot.ly/plotly-latest.min.js"></script>
    <link href="https://cdn.jsdelivr.net/npm/tailwindcss@3.0.0/dist/tailwind.min.css" rel="stylesheet">
    <style>
        body { font-family: "Courier New", Courier, monospace; }
        .dark-mode { background-color: #1a202c; color: #a0aec0; }
    </style>
</head>
<body class="dark-mode">
    <div class="container mx-auto p-4">
        <h1 class="text-3xl font-bold mb-4">GoProwl Metrics Dashboard</h1>
        <div id="activeRequests" class="mb-4 p-4 border border-gray-700 rounded"></div>
        <div id="pagesProcessed" class="mb-4 p-4 border border-gray-700 rounded"></div>
        <div id="errors" class="mb-4 p-4 border border-gray-700 rounded"></div>
    </div>
    <script>
        const queries = [
            { target: 'activeRequests', query: 'goprowl_active_requests', placeholder: '-' },
            { target: 'pagesProcessed', query: 'goprowl_pages_processed_total', placeholder: '-' },
            { target: 'errors', query: 'goprowl_errors_total', placeholder: '-' }
        ];
        
        // Fetch metrics data
        function fetchMetrics(target, query, placeholder) {
            fetch('/api/v1/query?query=' + query)
                .then(response => response.json())
                .then(data => updateChart(target, data, placeholder))
                .catch(() => document.getElementById(target).innerHTML = "<strong>" + target + ":</strong> " + placeholder);
        }

        // Update charts with fetched data
        function updateChart(target, data, placeholder) {
            const result = data.data.result;
            if (result.length === 0) {
                document.getElementById(target).innerHTML = "<strong>" + target + ":</strong> " + placeholder;
                return;
            }
            const value = result[0].value[1];
            document.getElementById(target).innerHTML = "<strong>" + target + ":</strong> " + value;
        }

        // Periodically fetch metrics data
        setInterval(() => {
            queries.forEach(({ target, query, placeholder }) => fetchMetrics(target, query, placeholder));
        }, 5000);
    </script>
</body>
</html>
`

// RegisterDashboard registers the dashboard handler with the provided mux
func RegisterDashboard(mux *http.ServeMux) {
	mux.HandleFunc("/dashboard", serveDashboard)
}

// serveDashboard serves the dashboard page
func serveDashboard(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("dashboard").Parse(dashboardTemplate))
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}
