Chart.defaults.global.defaultFontFamily = '-apple-system,system-ui,BlinkMacSystemFont,"Segoe UI",Roboto,"Helvetica Neue",Arial,sans-serif';
Chart.defaults.global.defaultFontColor = '#292b2c';


$(document).ready(function() {

    let cpuInfo = null;
    let consumo = [];
    let contador = 0;
    const headers = new Headers();
    headers.append('Content-Type', 'application/json');
    var chartHtml = document.getElementById("myChartCpu").getContext("2d");


    var chartConfig = {
        labels: [],
        scaleOverride : true,
        scaleSteps : 10,
        scaleStepWidth : 100,
        scaleStartValue : 0,
        datasets: [
            {
                label: "CPU %",
                lineTension: 0.3,
                backgroundColor: "rgba(2,117,216,0.2)",
                borderColor: "rgba(2,117,216,1)",
                pointRadius: 5,
                pointBackgroundColor: "rgba(2,117,216,1)",
                pointBorderColor: "rgba(255,255,255,0.8)",
                pointHoverRadius: 5,
                pointHoverBackgroundColor: "rgba(2,117,216,1)",
                pointHitRadius: 50,
                pointBorderWidth: 2,
                data: [],
            }
        ],


    };

    var options = {

        scales: {
            xAxes: [{
                time: {
                    unit: 'Second'
                },
                gridLines: {
                    display: true
                }/*,
                ticks: {
                    maxTicksLimit: 10
                }*/
            }],
            yAxes: [{
                ticks: {
                    min: 0,
                    max: 100,
                    maxTicksLimit: 10
                },
                gridLines: {
                    color: "rgba(0, 0, 0, .125)",
                }
            }],
        },

        legend: {
            display: false
        }
    
    };


    var myLineChart = new Chart(chartHtml, {
        type: "line",
        data: chartConfig,
        options: options
    });

    const init = {
        method: 'GET',
        headers
    };

    $('#overlayCpu').fadeOut(5000,function(){
        $('#divCpu').fadeIn(1000);
    });


    getCPUInfo();
    setInterval(function(){
        getCPUInfo();
    }, 5000);

    function getCPUInfo(){
        fetch('http://localhost:8080/cpuPorcentaje', init)
            .then(response => response.json())
            .then(data => {
                cpuInfo = data
                // text is the response body
            })
            .catch((e) => {
                console.log("ERROR: " + e.toString());
            });


        setTimeout(function(){
            var cardProcs = document.getElementById("cardCpu");
            cardProcs.innerHTML =
                " <br>Porcentaje CPU: "+ cpuInfo.porcentaje +" MB</br>" ;

            contador++;
            addData(contador, cpuInfo.porcentaje)
        }, 5000);

    }

    function addData(label, data) {
        myLineChart.data.labels.push(label);
        myLineChart.data.datasets.forEach((dataset) => {
            dataset.data.push(data);
        });
        myLineChart.update();
    }
});

