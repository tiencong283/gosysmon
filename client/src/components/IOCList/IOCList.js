import React from "react"
import "./IOCList.css"
import $ from "jquery"
import {Link} from "react-router-dom";

const title = "IOC List - GoSysmon"
const endpoint = "/api/ioc"
const iocTypes = ["Hash", "IP", "Domain"]

class IOCList extends React.Component {
    constructor(props) {
        super(props)
        this.state = {
            iocList: [],
        }
    }

    componentDidMount() {
        document.title = title
        $.ajax({
            url: endpoint,
            dataType: "json",
            success: function (data) {
                this.setState({
                    iocList: data,
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
                        <th>Type</th>
                        <th>Indicator</th>
                        <th>Notes</th>
                    </tr>
                    </thead>
                    <tbody>
                    {
                        this.state.iocList.map(function (ioc) {
                            return (
                                <tr>
                                    <td>{ioc.Timestamp}</td>
                                    <td>{iocTypes[ioc.IOCType]}</td>
                                    <td><a href={ioc.ExternalUrl}>{ioc.Indicator}</a></td>
                                    <td><Link
                                        to={`/process?ProviderGuid=${ioc.ProviderGuid}&ProcessGuid=${ioc.ProcessGuid}`}>
                                        {ioc.Message}</Link></td>
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

export default IOCList