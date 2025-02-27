document.addEventListener('DOMContentLoaded', function() {
    fetch('/api/groups')
        .then(response => response.json())
        .then(groups => {
            const urlParams = new URLSearchParams(window.location.search);
            const initialGroup = urlParams.get('group') || '';

            const select = d3.select('#group');

            let opt = select.append('option')
                .attr('value', group)
                .attr('disabled', "disabled")
                .text('— Groups —');
            if (!initialGroup) {
                opt.attr('selected', "selected")
            }

            groups.data.forEach(group => {
                let opt = select.append('option')
                    .attr('value', group)
                    .text(group);
                if (group === initialGroup) {
                    opt.attr('selected', "selected")
                }
            });
        });
});
