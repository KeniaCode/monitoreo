Chart.defaults.global.defaultFontFamily = '-apple-system,system-ui,BlinkMacSystemFont,"Segoe UI",Roboto,"Helvetica Neue",Arial,sans-serif';
Chart.defaults.global.defaultFontColor = '#292b2c';


$(document).ready(function() {

    let memInfo = null;
    let contador = 0;
    const headers = new Headers();
    headers.append('Content-Type', 'application/json');
    var chartHtml = document.getElementById("myChartRam").getContext("2d");


    var chartConfig = {
        labels: [],
        scaleOverride : true,
        scaleSteps : 10,
        scaleStepWidth : 100,
        scaleStartValue : 0,
        datasets: [
            {
                label: "RAM",
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
                    min: 60,
                    max: 100,
                    stepSize: 5
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


    $('#overlayRam').fadeOut(5000,function(){
        $('#divRam').fadeIn(1000);
    });

    getRamInfo();
    setInterval(function(){
        getRamInfo();
    }, 5000);


    function getRamInfo(){
        fetch('http://localhost:8080/memoria', init)
            .then(response => response.json())
            .then(data => {
                memInfo = data
                // text is the response body
            })
            .catch((e) => {
                console.log("ERROR: " + e.toString());
            });


        setTimeout(function(){

            var cardProcs = document.getElementById("cardRam");
            cardProcs.innerHTML =
                " <br>Total de RAM: "+ memInfo.total +" MB</br>" +
                " <br>RAM Consumida: "+ memInfo.libre +" MB</br>" +
                " <br>RAM Libre: "+ memInfo.consumo +" MB</br>" +
                " <br>Porcentaje: "+ memInfo.porcentaje +" MB</br>" ;
            contador++;
            addData(contador, memInfo.porcentaje)
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

