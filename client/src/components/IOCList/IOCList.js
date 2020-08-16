import React from "react"
import "./IOCList.css"
import $ from "jquery"
import {Link} from "react-router-dom";
import PaginationNav from "../PaginationNav/PaginationNav";

const title = "IOC List - GoSysmon"
const endpoint = "/api/ioc"
const iocTypes = ["Hash", "IP", "Domain"]

class IOCList extends React.Component {
    constructor(props) {
        super(props)
        this.state = {
            viewIOCs: [],
            iocList: [],
	    iocListRaw: [],
	    searchList: [],
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

    // pagination
    getViewElements(pageIdx) {
        return this.getViewElementsFrom(pageIdx, this.state.iocList)
    }

    getViewElementsFrom(pageIdx, iocList) {
        let startIdx = pageIdx * this.state.paging.elementsPerPage
        let endIdx = (pageIdx + 1) * this.state.paging.elementsPerPage
        return iocList.slice(startIdx, endIdx)
    }

    handlePrevious(event) {
        event.preventDefault()
        let newPageIdx = this.state.paging.currentPageIdx - 1
        if (newPageIdx < 0) {
            newPageIdx = 0
        }
        this.setState({
            viewIOCs: this.getViewElements(newPageIdx),
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
            viewIOCs: this.getViewElements(newPageIdx),
            paging: {
                ...this.state.paging,
                currentPageIdx: newPageIdx
            }
        })
    }

    handleInput(e){
	var searchIOCs = this.state.iocListRaw.filter( function (ioc) {
			var isSearched = false;
			if (ioc.ExternalUrl.indexOf(e.target.value) > -1) 
				isSearched = true;
			else if (ioc.Indicator.indexOf(e.target.value) > -1) 
				isSearched = true; 
		      if (isSearched) 
			return ioc;
		    })
	var isSearch = e.target.value=="" ? "0" : "1"
        this.setState({ textValue: e.target.value,
		viewIOCs: this.getViewElementsFrom(0, e.target.value!=="" ? searchIOCs : this.state.iocListRaw),
		searched: isSearch,
		searchList: searchIOCs,
		paging: {
                        ...this.state.paging,
			currentPageIdx: 0,
                        numOfPages: Math.floor(searchIOCs.length % this.state.paging.elementsPerPage == 0) ? Math.floor(searchIOCs.length / this.state.paging.elementsPerPage) : Math.floor(searchIOCs.length / this.state.paging.elementsPerPage + 1)
                    }
        });
    }

    handleSortByTime(e){ 
	var sortList = this.state.searched == "0" ? this.state.iocListRaw : this.state.searchList
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
		iocList: this.state.iocListRaw,
		sortType: this.state.sortType == "desc" ? "asc" : "desc",
		viewIOCs: this.getViewElementsFrom(0, list),
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
                    iocListRaw: data.sort(function(a, b) {
			  if (a.Timestamp.toUpperCase() < b.Timestamp.toUpperCase()) {
			    return 1;
			  }
			  if (a.Timestamp.toUpperCase() > b.Timestamp.toUpperCase()) {
			    return -1;
			  }
			  return 0;
			}),
                    viewIOCs: this.getViewElementsFrom(this.state.paging.currentPageIdx, data.sort(function(a, b) {
			  if (a.Timestamp.toUpperCase() < b.Timestamp.toUpperCase()) {
			    return 1;
			  }
			  if (a.Timestamp.toUpperCase() > b.Timestamp.toUpperCase()) {
			    return -1;
			  }
			  return 0;
			})),
                    iocList: data.sort(function(a, b) {
			  if (a.Timestamp.toUpperCase() < b.Timestamp.toUpperCase()) {
			    return 1;
			  }
			  if (a.Timestamp.toUpperCase() > b.Timestamp.toUpperCase()) {
			    return -1;
			  }
			  return 0;
			}),
                    paging: {
                        ...this.state.paging,
                        numOfPages: Math.floor(data.length % this.state.paging.elementsPerPage == 0) ? Math.floor(data.length / this.state.paging.elementsPerPage) : Math.floor(data.length / this.state.paging.elementsPerPage + 1)
                    }
                })
            }.bind(this),
        })
    }

    render() {
        return (
            <div className="list-table-container">
		<input value={this.state.textValue} onInput={this.handleInput.bind(this)}/>
                <PaginationNav paging={this.state.paging} handlePrevious={this.handlePrevious}
                               handleNext={this.handleNext}/>

                <table className="list-table hover unstriped">
                    <thead>
                    <tr>
                        <th><button onClick={this.handleSortByTime.bind(this)}>Timestamp</button></th>
                        <th>Type</th>
                        <th>Indicator</th>
                        <th>Notes</th>
                    </tr>
                    </thead>
                    <tbody>
                    {
                        this.state.viewIOCs.map(function (ioc) {
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
