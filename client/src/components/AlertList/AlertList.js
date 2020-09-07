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
        this.defaultSortInfo = {
            sortType: 1,
            sortFieldName: "Timestamp"
        }
        this.state = {
            viewAlerts: [],
            alert: {},
            paging: {   // pagination
                currentPageIdx: 0,
                elementsPerPage: 20,
                numOfPages: 0,
            },
            searchedText: "",
            sortInfo: this.defaultSortInfo
        }
        this.handlePrevious = this.handlePrevious.bind(this)
        this.handleNext = this.handleNext.bind(this)
        this.handleUserInput = this.handleUserInput.bind(this)
        this.handleSortClick = this.handleSortClick.bind(this)
    }

    sortAlertList(sortInfo, alertList) {
        switch (sortInfo.sortFieldName) {
            case "Timestamp":
                if (sortInfo.sortType === 1) {
                    alertList.sort((first, second) => first.Timestamp.localeCompare(second.Timestamp))
                } else {
                    alertList.sort((first, second) => second.Timestamp.localeCompare(first.Timestamp))
                }
                break
            case "Hostname":
                if (sortInfo.sortType === 1) {
                    alertList.sort((first, second) => first.HostName.localeCompare(second.HostName))
                } else {
                    alertList.sort((first, second) => second.HostName.localeCompare(first.HostName))
                }
                break
            case "Process":
                alertList.sort((first, second) => {
                    let firstProc = `${first.ProcessId} - ${first.ProcessImage}`
                    let secondProc = `${second.ProcessId} - ${second.ProcessImage}`
                    if (sortInfo.sortType === 1) {
                        return firstProc.localeCompare(secondProc)
                    } else {
                        return secondProc.localeCompare(firstProc)
                    }
                })
                break
            case "Technique":
                alertList.sort((first, second) => {
                    let firstTech = `${first.Technique.Id} - ${first.Technique.Name}`
                    let secondTech = `${second.Technique.Id} - ${second.Technique.Name}`
                    if (sortInfo.sortType === 1) {
                        return firstTech.localeCompare(secondTech)
                    } else {
                        return secondTech.localeCompare(firstTech)
                    }
                })
                break
            case "Notes":
                if (sortInfo.sortType === 1) {
                    alertList.sort((first, second) => first.Message.localeCompare(second.Message))
                } else {
                    alertList.sort((first, second) => second.Message.localeCompare(first.Message))
                }
                break
        }
    }

    handleSortClick(sortInfo) {
        this.sortAlertList(sortInfo, this.filteredAlerts)
        this.state.paging.currentPageIdx = 0
        this.setState({
            sortInfo: sortInfo,
            viewAlerts: this.getViewElements(0),
            paging: this.state.paging
        })
    }

    updateFilteredAlert(searchedText) {
        this.filteredAlerts = this.rawViewAlerts.filter(rawViewAlert => {
            for (const prop in rawViewAlert) {
                if (prop !== "idx" && rawViewAlert.hasOwnProperty(prop) && rawViewAlert[prop].indexOf(searchedText) >= 0) {
                    return true
                }
            }
            return false
        }).map(rawViewAlert => this.allAlerts[rawViewAlert.idx])
    }

    handleUserInput(e) {
        let searchedText = e.target.value
        this.updateFilteredAlert(searchedText)
        this.setState({
            viewAlerts: this.getViewElementsFrom(0, this.filteredAlerts),
            searchedText: searchedText,
            paging: {
                ...this.state.paging,
                currentPageIdx: 0,
                numOfPages: Math.floor(this.filteredAlerts.length / this.state.paging.elementsPerPage) + 1
            }
        })
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
        return this.getViewElementsFrom(pageIdx, this.filteredAlerts)
    }

    getViewElementsFrom(pageIdx, objList) {
        let startIdx = pageIdx * this.state.paging.elementsPerPage
        let endIdx = (pageIdx + 1) * this.state.paging.elementsPerPage
        return objList.slice(startIdx, endIdx)
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
                this.allAlerts = data
                this.rawViewAlerts = this.allAlerts.map((alert, idx) => {
                    return {
                        idx: idx,
                        Timestamp: alert.Timestamp,
                        Hostname: alert.HostName,
                        Process: `${alert.ProcessId} - ${alert.ProcessImage}`,
                        Technique: `${alert.Technique.Id} - ${alert.Technique.Name}`,
                        Notes: alert.Message
                    }
                })
                this.updateFilteredAlert(this.state.searchedText)
                this.setState({
                    viewAlerts: this.getViewElementsFrom(0, this.filteredAlerts),
                    paging: {
                        ...this.state.paging,
                        numOfPages: Math.floor(this.filteredAlerts.length / this.state.paging.elementsPerPage) + 1
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
                <div className="search-bar">
                    <input type="text" onInput={this.handleUserInput}/>
                </div>
                <PaginationNav paging={this.state.paging} handlePrevious={this.handlePrevious}
                               handleNext={this.handleNext}/>
                <AlertContextModel alert={this.state.alert}/>
                <table className="common-table">
                    <thead>
                    <tr>
                        <TableHeaderItem sortInfo={this.state.sortInfo}
                                         onClick={this.handleSortClick}>Timestamp</TableHeaderItem>
                        <TableHeaderItem sortInfo={this.state.sortInfo}
                                         onClick={this.handleSortClick}>Hostname</TableHeaderItem>
                        <TableHeaderItem sortInfo={this.state.sortInfo}
                                         onClick={this.handleSortClick}>Process</TableHeaderItem>
                        <TableHeaderItem sortInfo={this.state.sortInfo}
                                         onClick={this.handleSortClick}>Technique</TableHeaderItem>
                        <TableHeaderItem sortInfo={this.state.sortInfo}
                                         onClick={this.handleSortClick}>Notes</TableHeaderItem>
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

class TableHeaderItem extends React.Component {
    constructor(props) {
        super(props)
        this.handleOnClick = this.handleOnClick.bind(this)
    }

    handleOnClick() {
        let sortType = this.props.sortInfo.sortType + 1
        if (sortType === 3) {
            sortType = 1
        }
        this.props.onClick({
            sortType: sortType,
            sortFieldName: this.props.children
        })
    }

    render() {
        if (this.props.children !== this.props.sortInfo.sortFieldName) {
            return <th onClick={this.handleOnClick}>{this.props.children}</th>
        }
        switch (this.props.sortInfo.sortType) {
            case 1:
                return <th onClick={this.handleOnClick}>{this.props.children} <i className="fi-arrow-up"></i></th>
                break
            case 2:
                return <th onClick={this.handleOnClick}>{this.props.children} <i className="fi-arrow-down"></i></th>
                break
        }
    }
}

export default AlertList