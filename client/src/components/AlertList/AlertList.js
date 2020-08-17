import React from "react"
import {Link} from "react-router-dom"
import $ from "jquery"
import AlertContextModel from "../AlertContextModel/AlertContextModel"
import PaginationNav from "../PaginationNav/PaginationNav"

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
            viewAlerts: [],
            alertList: [],
            alert: {},
            paging: {   // pagination
                currentPageIdx: 0,
                elementsPerPage: 20,
                numOfPages: 0,
            }
        }
        this.handlePrevious = this.handlePrevious.bind(this)
        this.handleNext = this.handleNext.bind(this)
    }

    handleOpenSideBar(idx, event) {
        event.preventDefault()
        $("#alert-context").toggle()
        this.setState({
            alert: this.state.viewAlerts[idx],
        })
    }

    // pagination
    getViewElements(pageIdx) {
        return this.getViewElementsFrom(pageIdx, this.state.alertList)
    }

    getViewElementsFrom(pageIdx, alertList) {
        let startIdx = pageIdx * this.state.paging.elementsPerPage
        let endIdx = (pageIdx + 1) * this.state.paging.elementsPerPage
        return alertList.slice(startIdx, endIdx)
    }

    handlePrevious(event) {
        event.preventDefault()
        let newPageIdx = this.state.paging.currentPageIdx - 1
        if (newPageIdx < 0) {
            newPageIdx = 0
        }
        this.setState({
            viewAlerts: this.getViewElements(newPageIdx),
            paging: {
                ...this.state.paging,
                currentPageIdx: newPageIdx
            }
        })
    }

    handleNext(event) {
        event.preventDefault()
        let newPageIdx = this.state.paging.currentPageIdx + 1
        if (newPageIdx >= this.state.paging.numOfPages) {
            newPageIdx = this.state.paging.numOfPages - 1
        }
        this.setState({
            viewAlerts: this.getViewElements(newPageIdx),
            paging: {
                ...this.state.paging,
                currentPageIdx: newPageIdx
            }
        })
    }

    componentDidMount() {
        document.title = title
        $.ajax({
            url: endpoint,
            dataType: "json",
            success: function (data) {
                this.setState({
                    viewAlerts: this.getViewElementsFrom(this.state.paging.currentPageIdx, data),
                    alertList: data,
                    paging: {
                        ...this.state.paging,
                        numOfPages: Math.floor(data.length / this.state.paging.elementsPerPage) + 1
                    }
                })
            }.bind(this),
        })
    }

    renderAlerts() {
        return this.state.viewAlerts.map((alert, idx) => {
            return (
                <tr key={idx}>
                    <td className="col-timestamp">{alert.Timestamp}</td>
                    <td>{alert.HostName}</td>
                    <td>
                        <Link to={alert.ProcRefUrl}>{alert.ProcessId} - {alert.ProcessImage}</Link>
                    </td>
                    <td>
                        <a
                            onClick={this.handleOpenSideBar.bind(this, idx)}>{alert.Technique.Id}
                            - {alert.Technique.Name}</a>
                    </td>
                    <td className="col-notes">{alert.Message}</td>
                </tr>
            )
        })
    }

    render() {
        return (
            <div className="inner-content-wrapper">
                <PaginationNav paging={this.state.paging} handlePrevious={this.handlePrevious}
                               handleNext={this.handleNext}/>
                <AlertContextModel alert={this.state.alert}/>
                <table className="common-table">
                    <thead>
                    <tr>
                        <th>Timestamp</th>
                        <th>Hostname</th>
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

export default AlertList