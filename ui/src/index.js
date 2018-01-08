import React from 'react';
import ReactDOM from 'react-dom';
import $ from 'jquery';
import './index.css';
import App from './App';
import EditFile from './Edit';

// For static testing
let mockID = "00000000-0000-0000-0000-000000000000";
let mockRootFolderList = [
    {name: "folder 1", state: "expanded", contents: [
            {name: "filename-long.txt", state: "file", id: mockID, contents: []},
        ]
    },
    {name: "folder 2", state: "expanded", contents: [
            {name: "folder 3", state: "collapsed", contents :[
                    {name: "file2.txt", state: "file", id: mockID, contents: []},
                    {name: "file3.txt", state: "file", id: mockID,contents: []}
                ]}
        ]
    }
];
let mockEditFile = {
    id: mockID,
    name: "foo.txt",
    content: "Hello world\n\nHere's some text!"
};

let route = function(isTest) {
    if (window.location.pathname.startsWith("/edit/")) {
        if (isTest) {
            ReactDOM.render(<EditFile file={mockEditFile}/>, document.getElementById('root'));
            return
        }

        let fileID = window.location.pathname.substring("/edit/".length);
        // Load the file we wish to edit
        $.post('/api/1/load', {fileID: fileID}, (fileRaw) => {
            ReactDOM.render(<EditFile file={JSON.parse(fileRaw)}/>, document.getElementById('root'))
        }).fail(function () {
            window.location.replace("/login.html")
        })
    } else {
        if (isTest) {
            ReactDOM.render(<App rootFolderList={mockRootFolderList}/>, document.getElementById('root'));
            return
        }

        // Read in the files.
        $.getJSON('/api/1/list', (data) => {
            ReactDOM.render(<App rootFolderList={data}/>, document.getElementById('root'));
        }).fail(function () {
            window.location.replace("/login.html")
        });
    }
};
route(false);
