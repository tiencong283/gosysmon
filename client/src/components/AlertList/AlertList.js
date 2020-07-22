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
            alert: {},
        }
    }

    handleOpenSideBar(idx) {
        $("#alert-context").toggle()
        this.setState({
            alert: this.state.alertList[idx],
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

    getProcessUrl(alert) {
        return `/process?ProviderGuid=${alert.ProviderGuid}&ProcessGuid=${alert.ProcessGuid}`
    }

    renderAlerts() {
        return this.state.alertList.map((alert, idx) => {
            return (
                <tr>
                    <td>{alert.Timestamp}</td>
                    <td>{alert.HostName}</td>
                    <td><Link to={this.getProcessUrl(alert)}>{alert.ProcessId} - {alert.ProcessImage}</Link></td>
                    <td><span className="clickable"
                              onClick={this.handleOpenSideBar.bind(this, idx)}>{alert.Technique.Id} - {alert.Technique.Name}</span>
                    </td>
                    <td>{alert.Message}</td>
                </tr>
            )
        })
    }

    render() {
        return (
            <div className="list-table-container">
                <SideNav alert={this.state.alert}/>
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
                        this.renderAlerts()
                    }
                    </tbody>
                </table>
            </div>
        )
    }
}

class SideNav extends React.Component {
    renderHeader() {
        let alert = this.props.alert
        if ($.isEmptyObject(alert)) {
            return
        }
        return (
            <div className="alert-context-header"><a href={alert.Technique.Url}>Mitre
                ATT&CK <i className="fa fa-external-link"/></a></div>
        )
    }

    renderPropList() {
        let alert = this.props.alert
        if ($.isEmptyObject(alert)) {
            return
        }
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
        return properties.map(function (arr) {
            return (
                <tr>
                    <td>{arr[0]}</td>
                    <td>{arr[1]}</td>
                </tr>
            )
        })
    }

    render() {
        return (
            <div id="alert-context" className="sidenav">
                {
                    this.renderHeader()
                }
                <div className="alert-context-content">
                    <table>
                        <thead>
                        <th width="100">Property</th>
                        <th>Value</th>
                        </thead>
                        <tbody>
                        {
                            this.renderPropList()
                        }
                        </tbody>
                    </table>
                </div>
            </div>
        )
    }
}

export default AlertList