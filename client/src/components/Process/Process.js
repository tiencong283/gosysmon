import React from "react"
import "./Process.css"
import {useLocation} from "react-router-dom"
import $ from "jquery"

import ProcessTabLogo from "./gear.svg"
import FileTabLogo from "./file-earmark-binary.svg"
import ActTabLogo from "./camera-reels.svg"
import RelationshipTabLogo from "./diagram-3.svg"

import Highcharts from 'highcharts'
import HighchartsReact from 'highcharts-react-official'

require('highcharts/modules/sankey')(Highcharts)
require('highcharts/modules/organization')(Highcharts)
require('highcharts/modules/exporting')(Highcharts)
require('highcharts/modules/accessibility')(Highcharts)

const title = "Process Information - Gosysmon"
const processAPI = "/api/process"
const processRelAPI = "/api/process-tree"

// A custom hook that builds on useLocation to parse
// the query string for you.
function useQuery() {
    return new URLSearchParams(useLocation().search)
}

export default function ProcessWrapper() {
    let query = useQuery()
    return (
        <Process providerGuid={query.get("ProviderGuid")} processGuid={query.get("ProcessGuid")}/>
    )
}

class Process extends React.Component {
    constructor(props) {
        super(props)
        this.state = {
            proc: {},
            procRel: {},
            tabSegment: "#execution-details"
        }
        this.handleSwitchTab = this.handleSwitchTab.bind(this)
    }

    handleSwitchTab(e) {
        let segment = e.currentTarget.getAttribute("href")
        if (segment === "#relationship") {
            $.post({
                url: processRelAPI,
                async: false,
                dataType: "json",
                data: {
                    ProviderGuid: this.props.providerGuid,
                    ProcessGuid: this.props.processGuid
                },
                success: function (data) {
                    this.setState({
                        procRel: data,
                    })
                }.bind(this),
            })
        }
        this.setState({
            tabSegment: segment
        })
    }

    componentDidMount() {
        document.title = title
        $.post({
            url: processAPI,
            dataType: "json",
            data: {
                ProviderGuid: this.props.providerGuid,
                ProcessGuid: this.props.processGuid
            },
            success: function (data) {
                this.setState({
                    proc: data,
                })
            }.bind(this),
        })
    }

    render() {
        return (
            <div className="process-wrapper">
                <header className="process-header">
                    <ul>
                        <li><a href="#execution-details" onClick={this.handleSwitchTab}><img src={ProcessTabLogo}
                                                                                             alt=""/><span>Execution</span></a>
                        </li>
                        <li><a href="#file-defails" onClick={this.handleSwitchTab}><img src={FileTabLogo}
                                                                                        alt=""/><span>File</span></a>
                        </li>
                        <li><a href="#activity" onClick={this.handleSwitchTab}><img src={ActTabLogo}
                                                                                    alt=""/><span>Activities</span></a>
                        </li>
                        <li><a href="#relationship" onClick={this.handleSwitchTab}><img src={RelationshipTabLogo}
                                                                                        alt=""/><span>Relationship</span></a>
                        </li>
                    </ul>
                </header>
                <div className="process-content">
                    {this.state.tabSegment === "#execution-details" && <ProcessExecution proc={this.state.proc}/>}
                    {this.state.tabSegment === "#file-defails" && <ProcessImageFile proc={this.state.proc}/>}
                    {this.state.tabSegment === "#activity" && <ProcessActivities/>}
                    {this.state.tabSegment === "#relationship" && <ProcessRel procRel={this.state.procRel}/>}
                </div>
            </div>
        )
    }
}

const executionProps = [
    ["Process ID:", "ProcessId"],
    ["Image:", "Image"],
    ["Commandline:", "CommandLine"],
    ["CurrentDirectory:", "CurrentDirectory"],
    ["State:", "State"],
    ["Execution time:", "CreatedAt"],
    ["Stopped At:", "TerminatedAt"],
    ["Integrity Level:", "IntegrityLevel"],
]
const procStates = ["Running", "Stopped"]

class ProcessExecution extends React.Component {
    render() {
        let proc = this.props.proc
        proc.State = procStates[proc.State]
        return (
            <div>
                {
                    executionProps.map(function (prop) {
                        return <p><span class="pinfo-key">{prop[0]}</span><span>{proc[prop[1]]}</span></p>
                    })
                }
            </div>
        )
    }
}

const fileProps = [
    ["SHA1:", "SHA1"],
    ["SHA256:", "SHA256"],
    ["MD5:", "MD5"],
    ["OriginalFileName:", "OriginalFileName"],
    ["FileVersion:", "FileVersion"],
    ["Description:", "CreatedAt"],
    ["Product:", "Product"],
    ["Company:", "Company"],
]

class ProcessImageFile extends React.Component {
    render() {
        let proc = this.props.proc
        return (
            <div>
                {
                    fileProps.map(function (prop) {
                        return <p><span class="pinfo-key">{prop[0]}</span><span>{proc[prop[1]]}</span></p>
                    })
                }
            </div>
        )
    }
}

class ProcessActivities extends React.Component {
    render() {
        return <h3>ProcessActivities</h3>
    }
}

class ProcessRel extends React.Component {
    render() {
        let procTree = this.props.procRel
        console.log(procTree)

        const options = {
            chart: {
                inverted: true,
            },
            title: {
                text: 'Process Tree'
            },
            series: [{
                type: 'organization',
                name: '',
                keys: ['from', 'to'],
                data: procTree.Links,
                nodes: procTree.Nodes.map(function (node) {
                    return {
                        id: node.ProcessGuid,
                        name: node.ImageName,
                    }
                }),
                colorByPoint: false,
                color: 'white',
                dataLabels: {
                    color: 'black',
                    useHTML: false
                },
            }],
        }
        return (
            <div>
                <HighchartsReact highcharts={Highcharts} options={options}/>
            </div>
        )
    }
}