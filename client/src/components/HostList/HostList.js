import React from "react"
import "./HostList.css"
import $ from "jquery"

const endpoint = "/api/host"

class HostList extends React.Component {
    constructor(props) {
        super(props)
        this.state = {
            hostList: []
        }
    }

    componentDidMount() {
        $.ajax({
            url: endpoint,
            dataType: "json",
            success: function (data) {
                this.setState({
                    hostList: data,
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
                        <th>Host Name</th>
                        <th>Status</th>
                        <th>Joined At</th>
                    </tr>
                    </thead>
                    <tbody>
                    {
                        this.state.hostList.map(function (host) {
                            return (
                                <tr>
                                    <td><span>{host.Name}</span></td>
                                    <td><span>{host.Active ? "Active" : "Not active"}</span></td>
                                    <td><span>{host.FirstSeen}</span></td>
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

export default HostList