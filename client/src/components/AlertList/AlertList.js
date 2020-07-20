import React from "react"
import "./AlertList.css"
import {Link} from "react-router-dom"
import $ from "jquery"

const endpoint = "/api/alert"

class AlertList extends React.Component {
    constructor(props) {
        super(props)
        this.state = {
            alertList: []
        }
    }

    componentDidMount() {
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
            <div className="alert-table-container">
                <table className="alert-table hover unstriped">
                    <thead>
                    <tr>
                        <th>Timestamp</th>
                        <th>Host Name</th>
                        <th>Process</th>
                        <th>Technique</th>
                        <th>Message</th>
                    </tr>
                    </thead>
                    <tbody>
                    {
                        this.state.alertList.map(function (alert) {
                            return (
                                <tr>
                                    <td><span>{alert.Timestamp}</span></td>
                                    <td><span>{alert.HostName}</span></td>
                                    <td>
                                        <span><Link to={`/process?ProviderGuid=${alert.ProviderGuid}&ProcessGuid=${alert.ProcessGuid}`}>
                                        {alert.ProcessId} - {alert.ProcessImage}</Link></span>
                                    </td>
                                    <td><span><a href={alert.Technique.Url}>{alert.Technique.Id} - {alert.Technique.Name}</a></span></td>
                                    <td><span>{alert.Message}</span></td>
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