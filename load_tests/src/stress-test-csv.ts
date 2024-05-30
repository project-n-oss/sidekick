export function getStressTestCsv(data: any) {
    let csv: string[] = ['scenario,avg,min,med,max,p(90),p(95)'];

    for (const metric in data['metrics']) {
        // Check if the metrics is a http_req_duration scenario
        if (metric.startsWith('http_req_duration') && metric.includes('scenario')) {
            const scenario = metric.split('{')[1].split('}')[0];
            const values = data['metrics'][metric]['values'];

            const csvRow: string[] = [
                scenario,
                values['avg'].toString(),
                values['min'].toString(),
                values['med'].toString(),
                values['max'].toString(),
                values['p(90)'].toString(),
                values['p(95)'].toString(),
            ];

            csv.push(csvRow.join(','));
        }
    }

    return csv.join('\n');
}
