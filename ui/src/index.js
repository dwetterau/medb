import React from 'react';
import ReactDOM from 'react-dom';
import $ from 'jquery';
import './index.css';
import App from './App';

// For static testing
/*
ReactDOM.render(<App rootFolderList={
    {name: "folder 1", state: "collapsed", contents: [
            {name: "file1.txt", state: "file", contents: []},
        ]
    },
    {name: "folder 2", state: "expanded", contents: [
            {name: "folder 3", state: "collapsed", contents :[
                    {name: "file2.txt", state: "file", contents: []},
                    {name: "file3.txt", state: "file", contents: []}
                ]}
        ]
    }
]}/>, document.getElementById('root'));
*/

// Read in the files.
$.getJSON('/api/1/list', (data) => {
    ReactDOM.render(<App rootFolderList={data}/>, document.getElementById('root'));
}).fail(function() {
    window.location.replace("/login.html")
});
