import React from "react"
import Highcharts from "highcharts"
import HighchartsReact from "highcharts-react-official"
import Moment from "moment"
import MomentTimezone from "moment-timezone"
import $ from "jquery"

require('highcharts/modules/accessibility')(Highcharts)

window.moment = Moment
MomentTimezone()

const title = "Home - GoSysmon"
const endpoint = "/ws/event-processing-rate"
const techStatsAPI = "api/technique-stats"

Highcharts.setOptions({
    time: {
        timezone: 'Asia/Ho_Chi_Minh'
    }
})

class Home extends React.Component {
    constructor(props) {
        super(props)

        this.evenRateTheshold = 100
        this.techThreshold = 10
        this.state = {
            eventRateChartOptions: {
                chart: {
                    width: null,
                    zoomType: 'x',
                },
                title: {
                    text: 'Event Processing Rate'
                },
                subtitle: {
                    text: 'Number of events per second'
                },
                xAxis: {
                    title: {
                        text: 'Timeline'
                    },
                    type: 'datetime'
                },
                yAxis: {
                    title: {
                        text: 'Number of events/second'
                    }
                },
                legend: {
                    enabled: false
                },
                plotOptions: {
                    area: {
                        marker: {
                            enabled: false
                        },
                        lineWidth: 1,
                        lineColor: Highcharts.getOptions().colors[1],
                        color: Highcharts.getOptions().colors[2],
                        threshold: null
                    }
                },
                series: [{
                    type: 'area',
                    name: 'Event Rate',
                    data: []
                }],
                credits: {
                    enabled: false
                }
            },

            techStatsChartOptions: {
                chart: {
                    type: 'pie'
                },
                title: {
                    text: 'Mitre Attack Technique Statistics'
                },
            }
        }
        this.handleEventRateSwitch = this.handleEventRateSwitch.bind(this)
        this.handleResetTechniqueStats = this.handleResetTechniqueStats.bind(this)
    }

    componentWillUnmount() {
        this.conn.close()
    }

    componentDidMount() {
        document.title = title

        if (window["WebSocket"]) {
            this.startWebsocket()
        } else {
            console.log("Warn: WebSocket is not supported")
        }
        this.loadTechStats()
    }

    afterChartCreated = (chart) => {
        this.internalChart = chart
    }

    startWebsocket() {
        let wsScheme = "ws"
        if (document.location.protocol === "https:") {
            wsScheme += "s"
        }
        let wsEndpoint = wsScheme + "://" + document.location.host + endpoint
        this.conn = new WebSocket(wsEndpoint)
        this.conn.onerror = event => {
            console.log("WebSocket error: ", event)
        }
        this.conn.onmessage = event => {
            let msg = JSON.parse(event.data)
            let shouldShift = false
            if (this.internalChart.series[0].data.length > this.evenRateTheshold) {
                shouldShift = true
            }
            this.internalChart.series[0].addPoint([msg.Timestamp, msg.EventRate], true, shouldShift)
        }
    }

    loadTechStats() {
        fetch(techStatsAPI, {cache: 'no-cache'}).then(response => {
            if (response.ok) {
                return response.json()
            }
            console.log(`Failed to get ${techStatsAPI}, ${response}`)
            return null
        }, networkError => console.log(`Failed to get ${techStatsAPI}, ${networkError}`)).then(jsonResponse => {
            if (jsonResponse && jsonResponse.Counts) {
                let numOfTechs = 0
                for (let e of jsonResponse.Counts) {
                    numOfTechs += e.Count
                }
                let techShares = jsonResponse.Counts.map(e => {
                    return {
                        id: e.Technique.Id,
                        name: e.Technique.Name,
                        y: e.Count * 100 / numOfTechs
                    }
                })
                this.setState({
                        techStatsChartOptions: {
                            chart: {
                                type: 'pie'
                            },
                            title: {
                                text: 'Mitre Attack Technique Statistics'
                            },
                            plotOptions: {
                                pie: {
                                    allowPointSelect: true,
                                    cursor: 'pointer',
                                    dataLabels: {
                                        enabled: true,
                                        format: '<b>{point.name}</b>: {point.percentage:.1f} %'
                                    }
                                }
                            },
                            series: [{
                                name: '',
                                tooltip: {
                                    headerFormat: '',
                                    pointFormat: '<b>{point.id} - {point.name}</b>: {point.percentage:.1f} %'
                                },
                                colorByPoint: true,
                                data: techShares,
                            }],
                            credits: {
                                enabled: false
                            }
                        }
                    }
                )
            }
        })
    }

    handleEventRateSwitch(event) {
        if ($(event.target).text() === "Stop") {
            this.conn.close()
            $("#event-rate-button").text("Start")
        } else {
            this.state.eventRateChartOptions.series[0].data = []
            this.startWebsocket()
            $("#event-rate-button").text("Stop")
        }
    }

    handleResetTechniqueStats(event) {
        this.loadTechStats()
    }

    render() {
        return (
            <div className="inner-content-wrapper grid-x">
                <div className="event-rate cell medium-6">
                    <button id="event-rate-button" className="hollow button small"
                            onClick={this.handleEventRateSwitch}>Stop
                    </button>
                    <div>
                        <HighchartsReact
                            highcharts={Highcharts}
                            options={this.state.eventRateChartOptions}
                            callback={this.afterChartCreated}
                        />
                    </div>
                </div>
                <div className="technique-stats cell medium-6">
                    <button id="technique-stats-button" className="hollow button small"
                            onClick={this.handleResetTechniqueStats}>Refresh
                    </button>
                    <div>
                        {<HighchartsReact
                            highcharts={Highcharts}
                            options={this.state.techStatsChartOptions}/>}
                    </div>
                </div>
            </div>
        )
    }
}

export default Home