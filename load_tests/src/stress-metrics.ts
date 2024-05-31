type MetricData = {
    type: string;
    contains: string;
    values: {
        avg: number;
        min: number;
        med: number;
        max: number;
        'p(90)': number;
        'p(95)': number;
    };
};

type Data = {
    [key: string]: MetricData;
};

export function getCSVStressMetrics(data: Data) {
    let rows: string[][] = [];

    for (const metric in data) {
        if (metric.startsWith('http_req_duration') && metric.includes('scenario')) {
            const scenarioTag = metric.split('{')[1].split('}')[0];
            const name = scenarioTag.split(':')[1];
            const values = data[metric].values;

            rows.push([
                name,
                values['avg'].toString(),
                values['min'].toString(),
                values['med'].toString(),
                values['max'].toString(),
                values['p(90)'].toString(),
                values['p(95)'].toString(),
            ]);
        }
    }

    rows.sort((a, b) => {
        const numA = Number(a[0].substring(2));
        const numB = Number(b[0].substring(2));
        return numA - numB;
    });
    rows.unshift(['scenario', 'avg', 'min', 'med', 'max', 'p(90)', 'p(95)']);

    return rows.map((row) => row.join(',')).join('\n');
}
