import React from "react"
import {useLocation} from "react-router-dom"
import AlertContextModel from "../AlertContextModel/AlertContextModel"
import Highcharts from 'highcharts'
import HighchartsReact from 'highcharts-react-official'
import $ from "jquery"

import "./Process.css"
import ProcessTabLogo from "./gear.svg"
import FileTabLogo from "./file-earmark-binary.svg"
import ActTabLogo from "./camera-reels.svg"
import RelationshipTabLogo from "./diagram-3.svg"
import UserTabLogo from "./user.svg"
import Header from '../Header/Header'

require('highcharts/modules/networkgraph')(Highcharts)

const axios = require('axios')
axios.defaults.headers.post['Content-Type'] = 'application/x-www-form-urlencoded'

const title = "Process Information - Gosysmon"
const processAPI = "/api/process"
const processRelAPI = "/api/process-tree"
const processActivitiesAPI = "/api/process-activities"

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
        tabSegment: "#session",
        logoSrc: UserTabLogo,
        name: "Session"
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
    },
]

class Process extends React.Component {
    constructor(props) {
        super(props)
        this.state = {
            proc: {},
            procRel: {},
            procActivities: {},
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
            url: processActivitiesAPI,
            data: formData,
            headers: {'Content-Type': 'multipart/form-data'}
        }).then(function (response) {
            this.setState({
                procActivities: response.data,
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
        return procNavItems.map((navItem, idx) => {
            let active = navItem.tabSegment === this.state.tabSegment ? "process-tab-active" : ""
            return (
                <li key={idx} className={active}><a href={navItem.tabSegment} onClick={this.handleSwitchTab}><img
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
                    {this.state.tabSegment === "#activity" &&
                    <ProcessActivities procActivities={this.state.procActivities}/>}
                    {this.state.tabSegment === "#relationship" &&
                    <ProcessRel proc={this.state.proc} procRel={this.state.procRel}/>}
                    {this.state.tabSegment === "#session" && <ProcessSession proc={this.state.proc}/>}
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
                    this.executionProps.map(function (prop, idx) {
                        return <p key={idx}><span className="pinfo-key">{prop[0]}</span><span>{proc[prop[1]]}</span></p>
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

    renderHashes() {
        let proc = this.props.proc
        if (!proc || !proc.Hashes) {
            return
        }
        return (
            <div>
                <p><span className="pinfo-key">MD5:</span><span>{proc.Hashes.MD5}</span></p>
                <p><span className="pinfo-key">SHA256:</span><span>{proc.Hashes.SHA256}</span></p>
                <p><span className="pinfo-key">SHA1:</span><span>{proc.Hashes.SHA1}</span></p>
            </div>
        )
    }

    render() {
        let proc = this.props.proc
        return (
            <div>
                {
                    this.renderHashes()
                }
                {
                    this.fileProps.map(function (prop, idx) {
                        return <p key={idx}><span className="pinfo-key">{prop[0]}</span><span>{proc[prop[1]]}</span></p>
                    })
                }
            </div>
        )
    }
}

class ProcessSession extends React.Component {
    constructor(props) {
        super(props);
        this.executionProps = [
            ["User:", "User"],
            ["LogonGuid:", "LogonGuid"],
            ["LogonId:", "LogonId"],
            ["TerminalSessionId:", "TerminalSessionId"],
        ]
    }

    render() {
        let proc = this.props.proc
        if (!proc || proc.Abandoned) {
            return <p>Unknown</p>
        }
        return (
            <div>
                {
                    this.executionProps.map(function (prop, idx) {
                        return <p key={idx}><span className="pinfo-key">{prop[0]}</span><span>{proc[prop[1]]}</span></p>
                    })
                }
            </div>
        )
    }
}

class ProcessActivities extends React.Component {
    constructor(props) {
        super(props)
        this.state = {
            viewObject: {}
        }

    }

    handleOpenSideBar(idx, event) {
        event.preventDefault()
        $("#alert-context").toggle()
        this.setState({
            viewObject: this.props.procActivities[idx],
        })
    }

    renderActivities() {
        return this.props.procActivities.map((entry, idx) => {
            return (
                <tr key={idx}>
                    <td className="col-timestamp"><span>{entry.Timestamp}</span></td>
                    <td><a
                        onClick={this.handleOpenSideBar.bind(this, idx)}>{entry.Technique.Id} - {entry.Technique.Name}</a>
                    </td>
                </tr>
            )
        })
    }

    render() {
        if (!this.props.procActivities) {
            return <p>Loading ...</p>
        }
        return (
            <div className="procActivities-content">
                <AlertContextModel alert={this.state.viewObject}/>
                <table className="common-table">
                    <thead>
                    <tr>
                        <th>Timeline</th>
                        <th>Activity</th>
                    </tr>
                    </thead>
                    <tbody>
                    {
                        this.renderActivities()
                    }
                    </tbody>
                </table>
            </div>
        )
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
        return Object.keys(this.nodeColors).map((key, idx) => {
            let nodeColor = this.nodeColors[key]
            let style = {
                backgroundColor: nodeColor.color
            }
            return <span key={idx}><span className="circle" style={style}/>{nodeColor.description}</span>
        })
    }

    render() {
        if ($.isEmptyObject(this.props.procRel) || $.isEmptyObject(this.props.proc)) {
            return
        }
        let proc = this.props.proc
        const networkGraphOptions = {
            chart: {
                height: 800
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
            credits: {
                enabled: false
            }
        }
        return (
	    <div className="grid-container full">
                <div className="grid-x grid-margin-x main-container">
                    <Header />
                    <div className="cell auto content-wrapper">
                        <div className="processtree-content">
			    <div className="node-note">
			        {
			            this.renderNodeNotes()
			        }
			    </div>
			    <HighchartsReact highcharts={Highcharts} options={networkGraphOptions}/>
		        </div>
                    </div>
                </div>
            </div>
            
        )
    }
}
