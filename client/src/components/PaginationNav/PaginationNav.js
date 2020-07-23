import React from "react"
import "./PaginationNav.css"

class PaginationNav extends React.Component {
    hasPrevious() {
        return this.props.paging.currentPageIdx > 0
    }

    hasNext() {
        let paging = this.props.paging
        return paging.currentPageIdx < paging.numOfPages - 1
    }

    render() {
        return (
            <nav aria-label="Pagination">
                <ul className="pagination text-right">
                    {
                        this.hasPrevious() ? (
                            <li className="pagination-previous"><a href="#" aria-label="Previous page"
                                                                   onClick={this.props.handlePrevious}>Previous</a></li>
                        ) : (
                            <li className="pagination-previous disabled">Previous</li>
                        )
                    }
                    <li>{this.props.paging.numOfPages > 1 && `[${this.props.paging.currentPageIdx + 1} / ${this.props.paging.numOfPages}]`}</li>
                    {
                        this.hasNext() ? (
                            <li className="pagination-next"><a href="#" aria-label="Next page"
                                                               onClick={this.props.handleNext}>Next</a></li>
                        ) : (
                            <li className="pagination-next disabled">Next</li>
                        )
                    }
                </ul>
            </nav>
        )
    }
}

export default PaginationNav