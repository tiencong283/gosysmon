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
        var that = this
        $.ajax({
            url: endpoint,
            dataType: "json",
            success: function (data) {
                let tmpHostList = []
                $.each(data, function (index) {
                    console.log("host: ", data[index])
                    let host = data[index]
                    tmpHostList.push(
                        <tr>
                            <td><span>{host.Name}</span></td>
                            <td><span>{host.Active ? "Active" : "Not active"}</span></td>
                            <td><span>{host.FirstSeen}</span></td>
                        </tr>
                    )
                });
                console.log("number of hosts: ", tmpHostList.length)
                that.setState({
                    hostList: tmpHostList,
                })
            },
        })
    }

    render() {
        return (
            <div className="alert-table-container">
                <table className="alert-table hover unstriped">
                    <thead>
                    <tr>
                        <th>Computer Name</th>
                        <th>Status</th>
                        <th>Joined At</th>
                    </tr>
                    </thead>
                    <tbody>
                    {this.state.hostList}
                    </tbody>
                </table>
            </div>
        )
    }
}

export default HostList