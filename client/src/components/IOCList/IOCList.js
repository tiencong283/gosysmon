import React from "react"
import "./IOCList.css"
import $ from "jquery";

const endpoint = "/api/ioc"

const iocTypes = ["Hash", "IP", "Domain"]

class IOCList extends React.Component {
    constructor(props) {
        super(props)
        this.state = {
            iocList: []
        }
    }

    componentDidMount() {
        var that = this
        $.ajax({
            url: endpoint,
            dataType: "json",
            success: function (data) {
                let tmpIOCList = []
                $.each(data, function (index) {
                    console.log("ioc: ", data[index])
                    let ioc = data[index]
                    tmpIOCList.push(
                        <tr>
                            <td><span>{iocTypes[ioc.IOCType]}</span></td>
                            <td><span>{ioc.Indicator}</span></td>
                            <td><span>{ioc.Message}</span></td>
                            <td><a href={ioc.ExternalUrl}>Check</a></td>
                        </tr>
                    )
                });
                console.log("number of iocs: ", tmpIOCList.length)
                that.setState({
                    iocList: tmpIOCList,
                })
            },
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
                        <th>Virustotal</th>
                    </tr>
                    </thead>
                    <tbody>
                    {this.state.iocList}
                    </tbody>
                </table>
            </div>
        )
    }
}

export default IOCList