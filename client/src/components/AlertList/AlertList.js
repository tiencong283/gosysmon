import React from "react"
import "./AlertList.css"
import {Link} from "react-router-dom"
import $ from "jquery"
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
            searchList: [],
            alertListRaw: [],
            alertList: [],
            alert: {},
	    sortType: "desc",
	    searched: "0",
            paging: {   // pagination
                currentPageIdx: 0,
                elementsPerPage: 20,
                numOfPages: 0,
            }
        }
        this.handlePrevious = this.handlePrevious.bind(this)
        this.handleNext = this.handleNext.bind(this)
    }

    handleOpenSideBar(idx) {
        $("#alert-context").toggle()
        this.setState({
            alert: this.state.alertList[idx],
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

    handleInput(e){
	var searchAlerts = this.state.alertListRaw.filter( function (alert) {
			var isSearched = false;
			if (alert.ProcessImage.indexOf(e.target.value) > -1) 
				isSearched = true;
			else if (alert.HostName.indexOf(e.target.value) > -1) 
				isSearched = true; 
			else if (alert.Technique.Id.indexOf(e.target.value) > -1) 
				isSearched = true; 
			else if (alert.Technique.Name.indexOf(e.target.value) > -1) 
				isSearched = true; 
		      if (isSearched) 
			return alert;
		    })
	var isSearch = e.target.value=="" ? "0" : "1"
        this.setState({ textValue: e.target.value,
		viewAlerts: this.getViewElementsFrom(0, e.target.value!=="" ? searchAlerts : this.state.alertListRaw),
		searched: isSearch,
		searchList: searchAlerts,
		paging: {
                        ...this.state.paging,
			currentPageIdx: 0,
                        numOfPages: Math.floor(searchAlerts.length % this.state.paging.elementsPerPage == 0) ? Math.floor(searchAlerts.length / this.state.paging.elementsPerPage) : Math.floor(searchAlerts.length / this.state.paging.elementsPerPage + 1)
                    }
        });
    }

    handleSortByHostName(e){
	var sortList = this.state.searched == "0" ? this.state.alertListRaw : this.state.searchList 
	var list = this.state.sortType == "desc" 
		? sortList.sort(function(a, b) {
		  if (a.HostName.toUpperCase() < b.HostName.toUpperCase()) {
		    return -1;
		  }
		  if (a.HostName.toUpperCase() > b.HostName.toUpperCase()) {
		    return 1;
		  }
		  return 0;
		})
		: sortList.sort(function(a, b) {
		  if (a.HostName.toUpperCase() < b.HostName.toUpperCase()) {
		    return 1;
		  }
		  if (a.HostName.toUpperCase() > b.HostName.toUpperCase()) {
		    return -1;
		  }
		  return 0;
		})
        this.setState({ 
		alertList: this.state.alertListRaw,
		sortType: this.state.sortType == "desc" ? "asc" : "desc",
		viewAlerts: this.getViewElementsFrom(0, list),
		paging: {
			...this.state.paging,
			currentPageIdx: 0,
		}
        });
    }

    handleSortByTime(e){ 
	var sortList = this.state.searched == "0" ? this.state.alertListRaw : this.state.searchList
	var list = this.state.sortType == "desc" 
		? sortList.sort(function(a, b) {
		  if (a.Timestamp.toUpperCase() < b.Timestamp.toUpperCase()) {
		    return -1;
		  }
		  if (a.Timestamp.toUpperCase() > b.Timestamp.toUpperCase()) {
		    return 1;
		  }
		  return 0;
		})
		: sortList.sort(function(a, b) {
		  if (a.Timestamp.toUpperCase() < b.Timestamp.toUpperCase()) {
		    return 1;
		  }
		  if (a.Timestamp.toUpperCase() > b.Timestamp.toUpperCase()) {
		    return -1;
		  }
		  return 0;
		})
        this.setState({ 
		alertList: this.state.alertListRaw,
		sortType: this.state.sortType == "desc" ? "asc" : "desc",
		viewAlerts: this.getViewElementsFrom(0, list),
		paging: {
			...this.state.paging,
			currentPageIdx: 0,
		}
        });
    }

    componentDidMount() {
        document.title = title
        $.ajax({
            url: endpoint,
            dataType: "json",
            success: function (data) {
                this.setState({
                    alertListRaw: data.sort(function(a, b) {
			  if (a.Timestamp.toUpperCase() < b.Timestamp.toUpperCase()) {
			    return 1;
			  }
			  if (a.Timestamp.toUpperCase() > b.Timestamp.toUpperCase()) {
			    return -1;
			  }
			  return 0;
			}),
                    alertList: data.sort(function(a, b) {
			  if (a.Timestamp.toUpperCase() < b.Timestamp.toUpperCase()) {
			    return 1;
			  }
			  if (a.Timestamp.toUpperCase() > b.Timestamp.toUpperCase()) {
			    return -1;
			  }
			  return 0;
			}),
		    viewAlerts: this.getViewElementsFrom(this.state.paging.currentPageIdx, data.sort(function(a, b) {
			  if (a.Timestamp.toUpperCase() < b.Timestamp.toUpperCase()) {
			    return 1;
			  }
			  if (a.Timestamp.toUpperCase() > b.Timestamp.toUpperCase()) {
			    return -1;
			  }
			  return 0;
			})),
                    paging: {
                        ...this.state.paging,
                        numOfPages: Math.floor(data.length % this.state.paging.elementsPerPage == 0) ? Math.floor(data.length / this.state.paging.elementsPerPage) : Math.floor(data.length / this.state.paging.elementsPerPage + 1)
                    }
                })
            }.bind(this),
        })
    }

    getProcessUrl(alert) {
        return `/process?ProviderGuid=${alert.ProviderGuid}&ProcessGuid=${alert.ProcessGuid}`
    }

    renderAlerts() {
        return this.state.viewAlerts.map((alert, idx) => {
            return (
                <tr>
                    <td>{alert.Timestamp}</td>
                    <td>{alert.HostName}</td>
                    <td><Link to={this.getProcessUrl(alert)}>{alert.ProcessId} - {alert.ProcessImage}</Link></td>
                    <td><span className="clickable"
                              onClick={this.handleOpenSideBar.bind(this, idx)}>{alert.Technique.Id} - {alert.Technique.Name}</span>
                    </td>
                    <td>{alert.Message}</td>
                </tr>
            )
        })
    }

    render() {
        return (
            <div className="list-table-container">
		<h1>{document.title}</h1>
		<input value={this.state.textValue} onInput={this.handleInput.bind(this)}/>
                <PaginationNav paging={this.state.paging} handlePrevious={this.handlePrevious}
                               handleNext={this.handleNext}/>
                <SideNav alert={this.state.alert}/>
                <table className="list-table">
                    <thead>
                    <tr>
                        <th><button onClick={this.handleSortByTime.bind(this)}>Timestamp</button></th>
                        <th><button onClick={this.handleSortByHostName.bind(this)}>Host Name</button></th>
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

class SideNav extends React.Component {
    renderHeader() {
        let alert = this.props.alert
        if ($.isEmptyObject(alert)) {
            return
        }
        return (
            <div className="alert-context-header"><a href={alert.Technique.Url}>Mitre
                ATT&CK <i className="fa fa-external-link"/></a></div>
        )
    }

    renderPropList() {
        let alert = this.props.alert
        if ($.isEmptyObject(alert)) {
            return
        }
        let properties = Object.keys(alert.Context).map(function (key) {
            return [key, alert.Context[key]]
        })
        properties.sort(function (a, b) {
            if (a[0] > b[0]) {
                return 1
            }
            if (a[0] < b[0]) {
                return -1
            }
            return 0
        })
        return properties.map(function (arr) {
            return (
                <tr>
                    <td>{arr[0]}</td>
                    <td>{arr[1]}</td>
                </tr>
            )
        })
    }

    render() {
        return (
            <div id="alert-context" className="sidenav">
                {
                    this.renderHeader()
                }
                <div className="alert-context-content">
                    <table>
                        <thead>
                        <tr>
                            <th width="100">Property</th>
                            <th>Value</th>
                        </tr>
                        </thead>
                        <tbody>
                        {
                            this.renderPropList()
                        }
                        </tbody>
                    </table>
                </div>
            </div>
        )
    }
}

export default AlertList
