import React from "react"
import "./AlertList.css"
import {Link} from "react-router-dom"
import $ from "jquery"

const title = "Alert List - GoSysmon"
const endpoint = "/api/alert"

class AlertList extends React.Component {
    constructor(props) {
        super(props)
        this.state = {
            alertList: []
        }
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
        return (
            <div className="list-table-container">
                <table className="list-table hover unstriped">
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
                        this.state.alertList.map(function (alert) {
                            return (
                                <tr>
                                    <td>{alert.Timestamp}</td>
                                    <td>{alert.HostName}</td>
                                    <td><Link
                                        to={`/process?ProviderGuid=${alert.ProviderGuid}&ProcessGuid=${alert.ProcessGuid}`}>
                                        {alert.ProcessId} - {alert.ProcessImage}</Link>
                                    </td>
                                    <td><a
                                        href={alert.Technique.Url}>{alert.Technique.Id} - {alert.Technique.Name}</a>
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

export default AlertList