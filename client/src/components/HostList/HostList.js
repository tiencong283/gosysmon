import React from "react"
import "./HostList.css"
import $ from "jquery"

const title = "Client List - GoSysmon"
const endpoint = "/api/host"

class HostList extends React.Component {
    constructor(props) {
        super(props)
        this.state = {
            hostList: []
        }
    }

    componentDidMount() {
        document.title = title
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
            <div className="list-table-container">
                <table className="list-table hover unstriped">
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