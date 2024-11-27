package metrics

import (
	"html/template"
	"net/http"
)

const dashboardTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>GoProwl Metrics Dashboard</title>
    <script src="https://cdn.plot.ly/plotly-latest.min.js"></script>
</head>
<body>
    <h1>GoProwl Metrics Dashboard</h1>
    <div id="activeRequests"></div>
    <div id="pagesProcessed"></div>
    <div id="errors"></div>
    <script>
        // Add Prometheus query visualizations
        const queries = [
            {
                target: 'activeRequests',
                query: 'goprowl_active_requests'
            },
            {
                target: 'pagesProcessed',
                query: 'goprowl_pages_processed_total'
            },
            {
                target: 'errors',
                query: 'goprowl_errors_total'
            }
        ];
        
        // Fetch and display metrics
        setInterval(() => {
            queries.forEach(({target, query}) => {
                fetch('/api/v1/query?query=' + query)
                    .then(response => response.json())
                    .then(data => {
                        // Update charts
                    });
            });
        }, 5000);
    </script>
</body>
</html>
`

func RegisterDashboard(mux *http.ServeMux) {
	mux.HandleFunc("/dashboard", serveDashboard)
}

func serveDashboard(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("dashboard").Parse(dashboardTemplate))
	tmpl.Execute(w, nil)
}
