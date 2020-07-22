import React from "react"
import "./AlertList.css"
import {Link} from "react-router-dom"
import $ from "jquery"

const title = "Alert List - GoSysmon"
const endpoint = "/api/alert"

$(document).click(function (event) {
    let $target = $(event.target)
    if (!$target.closest("#alert-context").length && $("#alert-context").is(":visible")) {
        $("#alert-context").hide()
    }
})

class AlertList extends React.Component {
    constructor(props) {
        super(props)
        this.state = {
            alertList: [],
            alertIdx: -1,
        }
        this.handleOpenSideBar = this.handleOpenSideBar.bind(this)
    }

    handleOpenSideBar(event) {
        $("#alert-context").toggle()
        this.setState({
            alertIdx: parseInt(event.currentTarget.getAttribute("id").slice("alert-".length)),
        })
    }

    componentDidMount() {
        document.title = title
        $.ajax({
            url: endpoint,
            dataType: "json",
            success: function (data) {
                this.setState({
                    alertList: data,
                })
            }.bind(this),
        })
    }

    render() {
        let handleOpenSideBar = this.handleOpenSideBar
        return (
            <div className="list-table-container">
                <SideNav alert={this.state.alertList[this.state.alertIdx]}/>
                <table className="list-table">
                    <thead>
                    <tr>
                        <th>Timestamp</th>
                        <th>Host Name</th>
                        <th>Process</th>
                        <th>Technique</th>
                        <th>Notes</th>
                    </tr>
                    </thead>
                    <tbody>
                    {
                        this.state.alertList.map(function (alert, index) {
                            return (
                                <tr>
                                    <td>{alert.Timestamp}</td>
                                    <td>{alert.HostName}</td>
                                    <td><Link
                                        to={`/process?ProviderGuid=${alert.ProviderGuid}&ProcessGuid=${alert.ProcessGuid}`}>
                                        {alert.ProcessId} - {alert.ProcessImage}</Link>
                                    </td>
                                    <td><a id={`alert-${index}`} href="#"
                                           onClick={handleOpenSideBar}>{alert.Technique.Id} - {alert.Technique.Name}</a>
                                    </td>
                                    <td>{alert.Message}</td>
                                </tr>
                            )
                        })
                    }
                    </tbody>
                </table>
            </div>
        )
    }
}

class SideNav extends React.Component {
    render() {
        if (!this.props.alert) {
            return <span></span>
        }
        let alert = this.props.alert
        let properties = Object.keys(alert.Context).map(function (key) {
            return [key, alert.Context[key]]
        })
        properties.sort(function (a, b) {
            if (a[0] > b[0]) {
                return 1
            }
            if (a[0] < b[0]) {
                return -1
            }
            return 0
        })
        return (
            <div id="alert-context" className="sidenav">
                <div className="alert-context-header"><a href={alert.Technique.Url}>Mitre
                    ATT&CK <i className="fa fa-external-link"></i></a></div>
                <div className="alert-context-content">
                    <table>
                        <thead>
                        <th width="100">Property</th>
                        <th>Value</th>
                        </thead>
                        <tbody>
                        {
                            properties.map(function (arr) {
                                return (
                                    <tr>
                                        <td>{arr[0]}</td>
                                        <td>{arr[1]}</td>
                                    </tr>
                                )
                            })
                        }
                        </tbody>
                    </table>
                </div>
            </div>
        )
    }
}

export default AlertList