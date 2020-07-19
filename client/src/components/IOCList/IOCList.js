import React from "react"
import "./IOCList.css"
import $ from "jquery"

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
            <div className="ioc-table-container">
                <table className="ioc-table hover unstriped">
                    <thead>
                    <tr>
                        <th>Type</th>
                        <th>Indicator</th>
                        <th>Message</th>
                    </tr>
                    </thead>
                    <tbody>
                    {
                        this.state.iocList.map(function (ioc) {
                            return (
                                <tr>
                                    <td><span>{iocTypes[ioc.IOCType]}</span></td>
                                    <td><a href={ioc.ExternalUrl}>{ioc.Indicator}</a></td>
                                    <td><span>{ioc.Message}</span></td>
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