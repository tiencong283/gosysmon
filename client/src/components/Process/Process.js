import React from "react"
import "./Process.css"
import {useLocation} from "react-router-dom"

import ProcessTabLogo from "./gear.svg"
import FileTabLogo from "./file-earmark-binary.svg"
import ActTabLogo from "./camera-reels.svg"
import RelationshipTabLogo from "./diagram-3.svg"

import Highcharts from 'highcharts'
import HighchartsReact from 'highcharts-react-official'
import $ from "jquery"

require('highcharts/modules/sankey')(Highcharts)
require('highcharts/modules/networkgraph')(Highcharts)
require('highcharts/modules/exporting')(Highcharts)
require('highcharts/modules/accessibility')(Highcharts)

const axios = require('axios')
axios.defaults.headers.post['Content-Type'] = 'application/x-www-form-urlencoded'

const title = "Process Information - Gosysmon"
const processAPI = "/api/process"
const processRelAPI = "/api/process-tree"

// A custom hook that builds on useLocation to parse
// the query string for you. https://reactrouter.com/web/example/query-parameters
function useQuery() {
    return new URLSearchParams(useLocation().search)
}

export default function ProcessWrapper() {
    let query = useQuery()
    return (
        <Process hostId={query.get("HostId")} processGuid={query.get("ProcessGuid")}/>
    )
}

const procNavItems = [
    {
        tabSegment: "#execution-details",
        logoSrc: ProcessTabLogo,
        name: "Execution"
    },
    {
        tabSegment: "#file-defails",
        logoSrc: FileTabLogo,
        name: "File"
    },
    {
        tabSegment: "#activity",
        logoSrc: ActTabLogo,
        name: "Activities"
    },
    {
        tabSegment: "#relationship",
        logoSrc: RelationshipTabLogo,
        name: "Relationship"
    }
]

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
        e.preventDefault()
        this.setState({
            tabSegment: e.currentTarget.getAttribute("href"),
        })
    }

    componentDidMount() {
        document.title = title

        let formData = new FormData()
        formData.set("HostId", this.props.hostId)
        formData.set("ProcessGuid", this.props.processGuid)

        axios({
            method: 'POST',
            url: processRelAPI,
            data: formData,
            headers: {'Content-Type': 'multipart/form-data'}
        }).then(function (response) {
            this.setState({
                procRel: response.data,
            })
        }.bind(this)).catch(function (error) {
            console.log(error)
        })

        axios({
            method: 'POST',
            url: processAPI,
            data: formData,
            headers: {'Content-Type': 'multipart/form-data'}
        }).then(function (response) {
            this.setState({
                proc: response.data,
            })
        }.bind(this)).catch(function (error) {
            console.log(error)
        })
    }

    renderProcNavItems() {
        return procNavItems.map((navItem) => {
            let active = navItem.tabSegment === this.state.tabSegment ? "process-tab-active" : ""
            return (
                <li className={active}><a href={navItem.tabSegment} onClick={this.handleSwitchTab}><img
                    src={navItem.logoSrc}
                    alt=""/><span>{navItem.name}</span></a>
                </li>
            )
        })
    }

    render() {
        return (
            <div className="process-wrapper">
                <header className="process-header">
                    <ul>
                        {
                            this.renderProcNavItems()
                        }
                    </ul>
                </header>
                <div className="process-content">
                    {this.state.tabSegment === "#execution-details" && <ProcessExecution proc={this.state.proc}/>}
                    {this.state.tabSegment === "#file-defails" && <ProcessImageFile proc={this.state.proc}/>}
                    {this.state.tabSegment === "#activity" && <ProcessActivities/>}
                    {this.state.tabSegment === "#relationship" &&
                    <ProcessRel proc={this.state.proc} procRel={this.state.procRel}/>}
                </div>
            </div>
        )
    }
}

class ProcessExecution extends React.Component {
    constructor(props) {
        super(props);
        this.executionProps = [
            ["Process ID:", "ProcessId"],
            ["Image:", "Image"],
            ["Commandline:", "CommandLine"],
            ["CurrentDirectory:", "CurrentDirectory"],
            ["State:", "State"],
            ["Execution time:", "CreatedAt"],
            ["Stopped At:", "TerminatedAt"],
            ["Integrity Level:", "IntegrityLevel"],
        ]
    }

    render() {
        let proc = this.props.proc
        return (
            <div>
                {
                    this.executionProps.map(function (prop) {
                        return <p><span className="pinfo-key">{prop[0]}</span><span>{proc[prop[1]]}</span></p>
                    })
                }
            </div>
        )
    }
}

class ProcessImageFile extends React.Component {
    constructor(props) {
        super(props);
        this.fileProps = [
            ["OriginalFileName:", "OriginalFileName"],
            ["FileVersion:", "FileVersion"],
            ["Description:", "CreatedAt"],
            ["Product:", "Product"],
            ["Company:", "Company"],
        ]
    }

    render() {
        let proc = this.props.proc
        return (
            <div>
                <p><span className="pinfo-key">MD5:</span><span>{proc.Hashes.MD5}</span></p>
                <p><span className="pinfo-key">SHA256:</span><span>{proc.Hashes.SHA256}</span></p>
                <p><span className="pinfo-key">SHA1:</span><span>{proc.Hashes.SHA1}</span></p>
                {
                    this.fileProps.map(function (prop) {
                        return <p><span className="pinfo-key">{prop[0]}</span><span>{proc[prop[1]]}</span></p>
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
    constructor(props) {
        super(props)
        this.nodeColors = {
            child: {
                color: 'darkgray',
                description: 'Child Process'
            },
            ancestor: {
                color: 'brown',
                description: 'Ancestor Process'
            },
            focus: {
                color: 'cyan',
                description: 'Focused Process'
            }
        }
    }

    renderNodeNotes() {
        return Object.keys(this.nodeColors).map(key => {
            let nodeColor = this.nodeColors[key]
            let style = {
                backgroundColor: nodeColor.color
            }
            return <span><span className="circle" style={style}/>{nodeColor.description}</span>
        })
    }

    render() {
        if ($.isEmptyObject(this.props.procRel) || $.isEmptyObject(this.props.proc)) {
            return
        }
        let proc = this.props.proc
        const networkGraphOptions = {
            chart: {
                height: '100%'
            },
            title: {
                text: `Process Tree For ${proc.ImageName}`
            },
            subtitle: {
                text: `ProcessID: ${proc.ProcessId}, Image: ${proc.Image}`
            },
            tooltip: {
                formatter: function () {
                    return `<div><span>ProcessID: ${this.point.processId}</span><br>Image: ${this.point.image}</span></div>`
                }
            },
            series: [{
                type: 'networkgraph',
                name: '',
                layoutAlgorithm: {
                    enableSimulation: true,
                },
                turboThreshold: 0,
                marker: {
                    radius: 13
                },
                draggable: true,
                dataLabels: {
                    enabled: true,
                    format: '{point.imageName}',
                    linkFormat: '\u2192',
                    allowOverlap: true
                },
                keys: ['from', 'to'],
                data: this.props.procRel.Links,
                nodes: this.props.procRel.Nodes.map((node) => {
                    return {
                        id: node.ProcessGuid,
                        imageName: node.ImageName,
                        image: node.Image,
                        processId: node.ProcessId,
                        color: this.nodeColors[node.NodeType].color
                    }
                }),
            }],
            exporting: {
                enabled: false
            },
        }

        return (
            <div className="processtree-content">
                <div className="node-note">
                    {
                        this.renderNodeNotes()
                    }
                </div>
                <HighchartsReact highcharts={Highcharts} options={networkGraphOptions}/>
            </div>
        )
    }
}